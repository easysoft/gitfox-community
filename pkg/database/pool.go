// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package database

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// RewritingConnPool 包装 *sql.DB，在 SQL 执行前调用 SQLRewriter 重写 SQL。
// 仅用于需要方言重写的数据库（如 GaussDB），其他数据库不包装，保持原样执行。
//
// 通过实现 gorm.ConnPoolBeginner 返回包装了 *sql.Tx 的重写池，
// 确保事务内的查询同样经过 SQL 重写，否则 GORM 事务会将 ConnPool 替换为
// 裸 *sql.Tx，导致事务内驼峰列名查询不生效。
type RewritingConnPool struct {
	DB       *sql.DB
	Rewriter SQLRewriter
}

var (
	_ gorm.ConnPool         = (*RewritingConnPool)(nil)
	_ gorm.ConnPoolBeginner = (*RewritingConnPool)(nil)
	_ gorm.GetDBConnector   = (*RewritingConnPool)(nil)
)

// PrepareContext 重写 SQL 后准备语句。
func (p *RewritingConnPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return p.DB.PrepareContext(ctx, p.Rewriter.RewriteSQL(query))
}

// QueryContext 重写 SQL 后执行查询。
func (p *RewritingConnPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.DB.QueryContext(ctx, p.Rewriter.RewriteSQL(query), args...)
}

// QueryRowContext 重写 SQL 后执行单行查询。
func (p *RewritingConnPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.DB.QueryRowContext(ctx, p.Rewriter.RewriteSQL(query), args...)
}

// ExecContext 重写 SQL 后执行。
func (p *RewritingConnPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.DB.ExecContext(ctx, p.Rewriter.RewriteSQL(query), args...)
}

// BeginTx 开启事务（实现 gorm.ConnPoolBeginner）。
// 返回包装了 *sql.Tx 的重写池，使事务内查询也经过 SQL 重写。
func (p *RewritingConnPool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	tx, err := p.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &RewritingTxConnPool{tx: tx, Rewriter: p.Rewriter}, nil
}

// GetDBConn 返回底层 *sql.DB 连接（实现 gorm.GetDBConnector）。
func (p *RewritingConnPool) GetDBConn() (*sql.DB, error) {
	return p.DB, nil
}

// RewritingTxConnPool 包装 *sql.Tx，在 SQL 执行前调用 SQLRewriter 重写 SQL。
// 事务内查询通过该包装池执行，保证驼峰列名引号重写在事务中同样生效。
type RewritingTxConnPool struct {
	tx       *sql.Tx
	Rewriter SQLRewriter
}

var (
	_ gorm.ConnPool    = (*RewritingTxConnPool)(nil)
	_ gorm.TxCommitter = (*RewritingTxConnPool)(nil)
)

// PrepareContext 重写 SQL 后准备语句。
func (p *RewritingTxConnPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return p.tx.PrepareContext(ctx, p.Rewriter.RewriteSQL(query))
}

// QueryContext 重写 SQL 后执行查询。
func (p *RewritingTxConnPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.tx.QueryContext(ctx, p.Rewriter.RewriteSQL(query), args...)
}

// QueryRowContext 重写 SQL 后执行单行查询。
func (p *RewritingTxConnPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.tx.QueryRowContext(ctx, p.Rewriter.RewriteSQL(query), args...)
}

// ExecContext 重写 SQL 后执行。
func (p *RewritingTxConnPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.tx.ExecContext(ctx, p.Rewriter.RewriteSQL(query), args...)
}

// Commit 提交事务（实现 gorm.TxCommitter）。
func (p *RewritingTxConnPool) Commit() error {
	return p.tx.Commit()
}

// Rollback 回滚事务（实现 gorm.TxCommitter）。
func (p *RewritingTxConnPool) Rollback() error {
	return p.tx.Rollback()
}
