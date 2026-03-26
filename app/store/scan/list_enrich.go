package scan

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	"gorm.io/gorm"
)

type latestTaskJoin struct {
	ScanTask
	ExecutionStatus constant.CIStatus `gorm:"column:executionStatus"`
}

func applyLatestTaskToTrigger(m *scanModel.ScanTrigger, task *latestTaskJoin) {
	if m == nil || task == nil {
		return
	}
	taskID := int64(task.ID)
	m.LatestTaskID = &taskID
	if task.CreatedDate != nil {
		m.LatestTaskCreated = task.CreatedDate
	}
	m.LatestTaskResult = task.Result
	status := convertExecutionTaskStatus(task.ExecutionStatus)
	m.LatestTaskStatus = &status
}

func applyLatestTaskToPlan(m *scanModel.ScanPlan, task *latestTaskJoin) {
	if m == nil || task == nil {
		return
	}
	taskID := int64(task.ID)
	m.LatestTaskID = &taskID
	if task.CreatedDate != nil {
		m.LatestTaskCreated = task.CreatedDate
	}
	m.LatestTaskResult = task.Result
	status := convertExecutionTaskStatus(task.ExecutionStatus)
	m.LatestTaskStatus = &status
}

type repoMeta struct {
	Name    string
	SpaceID int64
}

func loadRepoMetaByRepoIDs(ctx context.Context, db *gorm.DB, repoIDs []uint64) (map[uint64]repoMeta, error) {
	repoMetaMap := make(map[uint64]repoMeta)
	if len(repoIDs) == 0 {
		return repoMetaMap, nil
	}

	type repoRow struct {
		ID      uint64 `gorm:"column:id"`
		Name    string `gorm:"column:name"`
		SpaceID uint64 `gorm:"column:spaceID"`
	}

	var repos []repoRow
	var repoStoreModel repo.Repository
	if err := db.WithContext(ctx).
		Table(repoStoreModel.TableName()).
		Select("id, name, spaceID").
		Where("id IN ? AND deleted = ?", repoIDs, 0).
		Find(&repos).Error; err != nil {
		return nil, err
	}
	for _, r := range repos {
		repoMetaMap[r.ID] = repoMeta{Name: r.Name, SpaceID: int64(r.SpaceID)}
	}
	return repoMetaMap, nil
}

func loadSpaceNamesBySpaceIDs(ctx context.Context, db *gorm.DB, spaceIDs []uint64) (map[uint64]string, error) {
	spaceNameMap := make(map[uint64]string)
	if len(spaceIDs) == 0 {
		return spaceNameMap, nil
	}

	type spaceRow struct {
		ID   uint64 `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}

	var spaces []spaceRow
	if err := db.WithContext(ctx).
		Table("ops_space").
		Select("id, name").
		Where("id IN ? AND deleted = ?", spaceIDs, 0).
		Find(&spaces).Error; err != nil {
		return nil, err
	}
	for _, s := range spaces {
		spaceNameMap[s.ID] = s.Name
	}
	return spaceNameMap, nil
}

func loadRepoSpaceIDsByRepoIDs(ctx context.Context, db *gorm.DB, repoIDs []uint64) (map[uint64]int64, error) {
	repoSpaceMap := make(map[uint64]int64)
	if len(repoIDs) == 0 {
		return repoSpaceMap, nil
	}

	type repoSpace struct {
		ID      uint64 `gorm:"column:id"`
		SpaceID uint64 `gorm:"column:spaceID"`
	}

	var repos []repoSpace
	var repoStoreModel repo.Repository
	if err := db.WithContext(ctx).
		Table(repoStoreModel.TableName()).
		Select("id, spaceID").
		Where("id IN ? AND deleted = ?", repoIDs, 0).
		Find(&repos).Error; err != nil {
		return nil, err
	}

	for _, repo := range repos {
		repoSpaceMap[repo.ID] = int64(repo.SpaceID)
	}
	return repoSpaceMap, nil
}

func loadTriggerIDsByPlanIDs(ctx context.Context, db *gorm.DB, planIDs []uint64) (map[uint64][]int64, error) {
	triggerMap := make(map[uint64][]int64)
	if len(planIDs) == 0 {
		return triggerMap, nil
	}

	var triggers []ScanPlanTrigger
	if err := db.WithContext(ctx).
		Where("planID IN ? AND deleted = ?", planIDs, false).
		Find(&triggers).Error; err != nil {
		return nil, err
	}

	for _, trigger := range triggers {
		triggerMap[trigger.PlanID] = append(triggerMap[trigger.PlanID], int64(trigger.ID))
	}
	return triggerMap, nil
}

func loadPlanConditionsByPlanIDs(ctx context.Context, db *gorm.DB, planIDs []uint64) (map[uint64][]int64, error) {
	conditionMap := make(map[uint64][]int64)
	if len(planIDs) == 0 {
		return conditionMap, nil
	}

	var conditions []ScanPlanCondition
	if err := db.WithContext(ctx).
		Where("planID IN ? AND triggerID = 0 AND deleted = ?", planIDs, false).
		Find(&conditions).Error; err != nil {
		return nil, err
	}

	for _, condition := range conditions {
		conditionMap[condition.PlanID] = append(conditionMap[condition.PlanID], int64(condition.ID))
	}
	return conditionMap, nil
}

func loadSolutionModelMap(ctx context.Context, db *gorm.DB, solutionIDs []uint64) (map[uint64]*scanModel.Solution, error) {
	solutionMap := make(map[uint64]*scanModel.Solution)
	if len(solutionIDs) == 0 {
		return solutionMap, nil
	}

	var solutions []SolutionRules
	if err := db.WithContext(ctx).
		Where("id IN ? AND deleted = ?", solutionIDs, false).
		Find(&solutions).Error; err != nil {
		return nil, err
	}

	solutionMap = make(map[uint64]*scanModel.Solution, len(solutions))
	for _, sol := range solutions {
		solutionMap[sol.ID] = mapSolutionToModel(sol)
	}
	return solutionMap, nil
}

func loadSolutionIDsByPlanIDs(ctx context.Context, db *gorm.DB, planIDs []uint64) (map[uint64][]int64, error) {
	solutionsMap := make(map[uint64][]int64)
	if len(planIDs) == 0 {
		return solutionsMap, nil
	}

	var bindings []ScanPlanSolution
	if err := db.WithContext(ctx).
		Where("planID IN ? AND deleted = ?", planIDs, false).
		Find(&bindings).Error; err != nil {
		return nil, err
	}

	for _, b := range bindings {
		solutionsMap[b.PlanID] = append(solutionsMap[b.PlanID], int64(b.SolutionID))
	}
	return solutionsMap, nil
}

func loadLatestTasksByPlanIDs(ctx context.Context, db *gorm.DB, planIDs []uint64) (map[uint64]*latestTaskJoin, error) {
	latestTaskMap := make(map[uint64]*latestTaskJoin)
	if len(planIDs) == 0 {
		return latestTaskMap, nil
	}

	subQuery := db.Model(&ScanTask{}).
		Select("MAX(id) as id").
		Where("planID IN ? AND deleted = ?", planIDs, false).
		Group("planID")

	var tasks []latestTaskJoin
	var scanTaskModel ScanTask
	if err := db.WithContext(ctx).
		Table(scanTaskModel.TableName()).
		Select("ops_scan_tasks.*, e.status as executionStatus").
		Joins("INNER JOIN ops_pipeline_executions e ON ops_scan_tasks.executionID = e.id").
		Where("ops_scan_tasks.id IN (?)", subQuery).
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	for i := range tasks {
		latestTaskMap[tasks[i].PlanID] = &tasks[i]
	}
	return latestTaskMap, nil
}

func loadTriggerConditionsByTriggerIDs(ctx context.Context, db *gorm.DB, triggerIDs []uint64) (map[uint64][]int64, map[uint64][]*scanModel.ScanCondition, error) {
	conditionsIDMap := make(map[uint64][]int64)
	conditionsObjMap := make(map[uint64][]*scanModel.ScanCondition)
	if len(triggerIDs) == 0 {
		return conditionsIDMap, conditionsObjMap, nil
	}

	var conditions []ScanPlanCondition
	if err := db.WithContext(ctx).
		Where("triggerID IN ? AND deleted = ?", triggerIDs, false).
		Find(&conditions).Error; err != nil {
		return nil, nil, err
	}

	for _, c := range conditions {
		conditionsIDMap[c.TriggerID] = append(conditionsIDMap[c.TriggerID], int64(c.ID))
		conditionsObjMap[c.TriggerID] = append(conditionsObjMap[c.TriggerID], mapScanConditionToModel(c))
	}
	return conditionsIDMap, conditionsObjMap, nil
}

func loadLatestTasksByTriggerIDs(ctx context.Context, db *gorm.DB, triggerIDs []uint64) (map[uint64]*latestTaskJoin, error) {
	latestTaskMap := make(map[uint64]*latestTaskJoin)
	if len(triggerIDs) == 0 {
		return latestTaskMap, nil
	}

	subQuery := db.Model(&ScanTask{}).
		Select("MAX(id) as id").
		Where("triggerID IN ? AND deleted = ?", triggerIDs, false).
		Group("triggerID")

	var tasks []latestTaskJoin
	var scanTaskModel ScanTask
	if err := db.WithContext(ctx).
		Table(scanTaskModel.TableName()).
		Select("ops_scan_tasks.*, e.status as executionStatus").
		Joins("INNER JOIN ops_pipeline_executions e ON ops_scan_tasks.executionID = e.id").
		Where("ops_scan_tasks.id IN (?)", subQuery).
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	for i := range tasks {
		if tasks[i].TriggerID != nil {
			latestTaskMap[*tasks[i].TriggerID] = &tasks[i]
		}
	}
	return latestTaskMap, nil
}

func loadTriggerSolutionsByTriggerIDs(ctx context.Context, db *gorm.DB, triggerIDs []uint64) (map[uint64][]int64, map[uint64][]*scanModel.Solution, error) {
	solutionsIDMap := make(map[uint64][]int64)
	solutionsObjMap := make(map[uint64][]*scanModel.Solution)
	if len(triggerIDs) == 0 {
		return solutionsIDMap, solutionsObjMap, nil
	}

	var triggerSolutions []ScanTriggerSolution
	if err := db.WithContext(ctx).
		Where("triggerID IN ?", triggerIDs).
		Find(&triggerSolutions).Error; err != nil {
		return nil, nil, err
	}
	if len(triggerSolutions) == 0 {
		return solutionsIDMap, solutionsObjMap, nil
	}

	allSolutionIDs := make([]uint64, 0, len(triggerSolutions))
	triggerToSolutionIDs := make(map[uint64][]uint64)
	for _, ts := range triggerSolutions {
		triggerToSolutionIDs[ts.TriggerID] = append(triggerToSolutionIDs[ts.TriggerID], ts.SolutionID)
		solutionsIDMap[ts.TriggerID] = append(solutionsIDMap[ts.TriggerID], int64(ts.SolutionID))
		allSolutionIDs = append(allSolutionIDs, ts.SolutionID)
	}

	solutionMap, err := loadSolutionModelMap(ctx, db, allSolutionIDs)
	if err != nil {
		return nil, nil, err
	}
	for tid, sids := range triggerToSolutionIDs {
		for _, sid := range sids {
			if obj, ok := solutionMap[sid]; ok {
				solutionsObjMap[tid] = append(solutionsObjMap[tid], obj)
			}
		}
	}
	return solutionsIDMap, solutionsObjMap, nil
}
