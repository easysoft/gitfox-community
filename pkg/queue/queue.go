package queue

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/easysoft/gitfox-community/pkg/database"
)

// 使用给定选项创建新队列。
func New(opts NewOpts) *Queue {
	if opts.DB == nil {
		panic("db cannot be nil")
	}

	if opts.Name == "" {
		panic("name cannot be empty")
	}

	if opts.MaxReceive < 0 {
		panic("max receive cannot be negative")
	}

	if opts.MaxReceive == 0 {
		opts.MaxReceive = 3
	}

	if opts.Timeout < 0 {
		panic("timeout cannot be negative")
	}

	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Second
	}

	if opts.SQLFlavor < SQLFlavorSQLite || opts.SQLFlavor > SQLFlavorMYSQL {
		panic("unsupported SQL flavor " + fmt.Sprint(opts.SQLFlavor))
	}

	return &Queue{
		db:         opts.DB,
		flavor:     opts.SQLFlavor,
		name:       opts.Name,
		maxReceive: opts.MaxReceive,
		timeout:    opts.Timeout,
	}
}

// 向队列发送消息
func (q *Queue) Send(ctx context.Context, m Message) error {
	return database.InTx(ctx, q.db, func(tx *sql.Tx) error {
		return q.SendTx(ctx, tx, m)
	})
}

// SendTX类似于Send，但在现有事务中。
func (q *Queue) SendTx(ctx context.Context, tx *sql.Tx, m Message) error {
	_, err := q.SendAndGetIDTx(ctx, tx, m)
	return err
}

// SendAndGetID 类似于 Send，但也返回消息ID，可以使用
// 与消息交互而不首先接收它。
func (q *Queue) SendAndGetID(ctx context.Context, m Message) (ID, error) {
	var id ID
	err := database.InTx(ctx, q.db, func(tx *sql.Tx) error {
		var err error
		id, err = q.SendAndGetIDTx(ctx, tx, m)
		return err
	})
	return id, err
}

// SendAndGetIDTx 类似 SendAndGetID，但在现有事务中。
func (q *Queue) SendAndGetIDTx(ctx context.Context, tx *sql.Tx, m Message) (ID, error) {
	if m.Delay < 0 {
		panic("delay cannot be negative")
	}

	timeout := time.Now().UTC().Add(m.Delay)

	var id ID
	switch q.flavor {
	case SQLFlavorSQLite:
		query := `insert into queue (queue, body, timeout, priority) values (?, ?, ?, ?) returning id`
		if err := tx.QueryRowContext(ctx, query, q.name, m.Body, timeout.Format(rfc3339Milli), m.Priority).Scan(&id); err != nil {
			return "", err
		}

	case SQLFlavorPostgreSQL:
		query := `insert into queue (queue, body, timeout, priority) values ($1, $2, $3, $4) returning id`
		if err := tx.QueryRowContext(ctx, query, q.name, m.Body, timeout, m.Priority).Scan(&id); err != nil {
			return "", err
		}
	}

	return id, nil
}

// Receive 从队列中接收消息，或 nil 如果没有消息。
func (q *Queue) Receive(ctx context.Context) (*Message, error) {
	var m *Message
	err := database.InTx(ctx, q.db, func(tx *sql.Tx) error {
		var err error
		m, err = q.ReceiveTx(ctx, tx)
		return err
	})
	return m, err
}

// ReceiveTx 从队列中接收消息，或 nil 如果没有消息。
func (q *Queue) ReceiveTx(ctx context.Context, tx *sql.Tx) (*Message, error) {
	now := time.Now().UTC()
	timeout := now.Add(q.timeout)

	var m Message

	switch q.flavor {
	case SQLFlavorSQLite:
		query := `
			update queue
			set
				timeout = ?,
				received = received + 1
			where id = (
				select id from queue
				where
					queue = ? and
					? >= timeout and
					received < ?
				order by priority desc, created
				limit 1
			)
			returning id, body`

		if err := tx.QueryRowContext(ctx, query, timeout.Format(rfc3339Milli), q.name, now.Format(rfc3339Milli), q.maxReceive).Scan(&m.ID, &m.Body); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}

	case SQLFlavorPostgreSQL:
		query := `
			update queue
			set
				timeout = $1,
				received = received + 1
			where id = (
				select id from queue
				where
					queue = $2 and
					$3 >= timeout and
					received < $4
				order by priority desc, created
				limit 1
			)
			returning id, body`

		if err := tx.QueryRowContext(ctx, query, timeout, q.name, now, q.maxReceive).Scan(&m.ID, &m.Body); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
	}

	return &m, nil
}

// ReceiveAndWait 从队列中接收消息，以给定的间隔轮询，直到上下文被取消。
// 如果上下文被取消，错误将是非空的。请参见[context.Context.Err]。
func (q *Queue) ReceiveAndWait(ctx context.Context, interval time.Duration) (*Message, error) {
	m, err := q.Receive(context.WithoutCancel(ctx))
	if err != nil {
		return nil, err
	}
	if m != nil {
		return m, nil
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			m, err := q.Receive(context.WithoutCancel(ctx))
			if err != nil {
				return nil, err
			}
			if m != nil {
				return m, nil
			}
		}
	}
}

// Extend 从现在开始将消息超时延长给定的延迟。
func (q *Queue) Extend(ctx context.Context, id ID, delay time.Duration) error {
	return database.InTx(ctx, q.db, func(tx *sql.Tx) error {
		return q.ExtendTx(ctx, tx, id, delay)
	})
}

// ExtendTx ExtendTx 类似于 Extend，但在现有事务中。
func (q *Queue) ExtendTx(ctx context.Context, tx *sql.Tx, id ID, delay time.Duration) error {
	if delay < 0 {
		panic("delay cannot be negative")
	}

	timeout := time.Now().UTC().Add(delay)

	var err error
	switch q.flavor {
	case SQLFlavorSQLite:
		_, err = tx.ExecContext(ctx, `update queue set timeout = ? where queue = ? and id = ?`, timeout.Format(rfc3339Milli), q.name, id)

	case SQLFlavorPostgreSQL:
		_, err = tx.ExecContext(ctx, `update queue set timeout = $1 where queue = $2 and id = $3`, timeout, q.name, id)
	}

	return err
}

// Delete 按id从队列中删除消息。
func (q *Queue) Delete(ctx context.Context, id ID) error {
	return database.InTx(ctx, q.db, func(tx *sql.Tx) error {
		return q.DeleteTx(ctx, tx, id)
	})
}

// DeleteTx DeleteTx类似于Delete，但在现有事务中。
func (q *Queue) DeleteTx(ctx context.Context, tx *sql.Tx, id ID) error {
	var err error
	switch q.flavor {
	case SQLFlavorSQLite:
		_, err = tx.ExecContext(ctx, `delete from queue where queue = ? and id = ?`, q.name, id)

	case SQLFlavorPostgreSQL:
		_, err = tx.ExecContext(ctx, `delete from queue where queue = $1 and id = $2`, q.name, id)
	}

	return err
}
