package scan

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ IScanTriggerSolution = (*ScanTriggerSolutionStore)(nil)

type ScanTriggerSolutionStore struct {
	db *gorm.DB
}

func NewScanTriggerSolutionStore(db *gorm.DB) *ScanTriggerSolutionStore {
	return &ScanTriggerSolutionStore{db: db}
}

// ReplaceBatch 先物理删除旧关联，再批量创建新关联。
// 使用物理删除是因为唯一索引 (triggerID, solutionID, deleted) 在软删模式下
// 多次操作同一组合会产生 Duplicate entry 冲突。
func (s *ScanTriggerSolutionStore) ReplaceBatch(ctx context.Context, triggerID uint64, solutionIDs []int64, createdBy string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 物理删除旧关联
		if err := tx.Where("triggerID = ?", triggerID).
			Delete(&ScanTriggerSolution{}).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		// 批量创建新关联
		for _, sid := range solutionIDs {
			record := &ScanTriggerSolution{
				TriggerID:  triggerID,
				SolutionID: uint64(sid),
				CreatedBy:  createdBy,
				EditedBy:   createdBy,
			}
			if err := tx.Create(record).Error; err != nil {
				return database.ProcessSQLErrorf(ctx, err)
			}
		}
		return nil
	})
}

// ListByTriggerIDs 批量查询多个触发器关联的 solutionID 列表，返回 map[triggerID][]solutionID。
func (s *ScanTriggerSolutionStore) ListByTriggerIDs(ctx context.Context, triggerIDs []uint64) (map[uint64][]int64, error) {
	if len(triggerIDs) == 0 {
		return make(map[uint64][]int64), nil
	}
	var records []ScanTriggerSolution
	if err := s.db.WithContext(ctx).
		Where("triggerID IN ?", triggerIDs).
		Find(&records).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	result := make(map[uint64][]int64, len(triggerIDs))
	for _, r := range records {
		result[r.TriggerID] = append(result[r.TriggerID], int64(r.SolutionID))
	}
	return result, nil
}

// DeleteByTriggerID 物理删除指定触发器的所有关联。
func (s *ScanTriggerSolutionStore) DeleteByTriggerID(ctx context.Context, triggerID uint64) error {
	if err := s.db.WithContext(ctx).
		Where("triggerID = ?", triggerID).
		Delete(&ScanTriggerSolution{}).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// ListByTriggerID 查询单个触发器关联的 solutionID 列表。
func (s *ScanTriggerSolutionStore) ListByTriggerID(ctx context.Context, triggerID uint64) ([]int64, error) {
	var records []ScanTriggerSolution
	if err := s.db.WithContext(ctx).
		Where("triggerID = ?", triggerID).
		Find(&records).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	ids := make([]int64, 0, len(records))
	for _, r := range records {
		ids = append(ids, int64(r.SolutionID))
	}
	return ids, nil
}
