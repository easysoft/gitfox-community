package scan

import (
	"context"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	execStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/execution"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	spaceStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newScanStoreTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(
		&spaceStore.Space{},
		&repoStore.Repository{},
		&ScanPlan{},
		&ScanPlanTrigger{},
		&ScanPlanCondition{},
		&ScanPlanSolution{},
		&ScanTriggerSolution{},
		&SolutionRules{},
		&ScanTask{},
		&execStore.Execution{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	return db
}

func TestScanTriggerStoreListEnrichment(t *testing.T) {
	t.Parallel()

	db := newScanStoreTestDB(t)
	store := NewScanTriggerStore(db)
	ctx := context.Background()
	now := datetime.NowLocalTime()
	result := constant.ScanTaskResultPass

	repo := repoStore.Repository{ID: 2, SpaceID: 99, Identifier: "repo-2", Deleted: 0}
	if err := db.WithContext(ctx).Create(&repo).Error; err != nil {
		t.Fatalf("seed repo failed: %v", err)
	}

	trigger := ScanPlanTrigger{ID: 11, PlanID: 1, RepoID: 2, Name: "trigger-1", Type: string(scanModel.ScanTriggerTypeEvent), ScanType: string(constant.ScanTypeIncremental), Actions: string(constant.TriggerActionBranchUpdated), CreatedBy: "tester", CreatedDate: now, EditedBy: "tester", EditedDate: now}
	if err := db.WithContext(ctx).Create(&trigger).Error; err != nil {
		t.Fatalf("seed trigger failed: %v", err)
	}
	condition := ScanPlanCondition{ID: 21, RepoID: 2, PlanID: 1, TriggerID: 11, Priority: constant.ScanRulePriorityLow, Type: constant.ScanRuleTypeDefect, Unit: constant.ScanRuleUnitCount, Threshold: 3, CreatedBy: "tester", CreatedDate: now, EditedBy: "tester", EditedDate: now}
	if err := db.WithContext(ctx).Create(&condition).Error; err != nil {
		t.Fatalf("seed condition failed: %v", err)
	}
	solution := SolutionRules{ID: 31, Name: "solution-1", Desc: "desc", Tag: "tag", Status: constant.ScanRuleStatusEnabled, CreatedBy: "tester", CreatedDate: now, EditedBy: "tester", EditedDate: now}
	if err := db.WithContext(ctx).Create(&solution).Error; err != nil {
		t.Fatalf("seed solution failed: %v", err)
	}
	triggerSolution := ScanTriggerSolution{ID: 41, TriggerID: 11, SolutionID: 31, CreatedBy: "tester", CreatedDate: now, EditedBy: "tester", EditedDate: now}
	if err := db.WithContext(ctx).Create(&triggerSolution).Error; err != nil {
		t.Fatalf("seed trigger solution failed: %v", err)
	}
	execution := execStore.Execution{ID: 51, PipelineID: 1, Parent: 0, Trigger: "manual", Commit: "abc", Ref: "refs/heads/main", Params: "{}", Status: string(constant.CIStatusSuccess), CreatedBy: "tester", CreatedDate: now, EditedDate: now}
	if err := db.WithContext(ctx).Create(&execution).Error; err != nil {
		t.Fatalf("seed execution failed: %v", err)
	}
	task := ScanTask{ID: 61, PlanID: 1, ExecutionID: 51, RepoNumber: 2, PlanNumber: 1, TriggerID: ptrUint64(11), Status: constant.ScanTaskStatusSuccess, Result: &result, CreatedBy: "tester", CreatedDate: now, EditedDate: now, ScanType: constant.ScanTypeIncremental}
	if err := db.WithContext(ctx).Create(&task).Error; err != nil {
		t.Fatalf("seed task failed: %v", err)
	}

	page, err := store.List(ctx, 2, 1, scanModel.ScanTriggerReqFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Data) != 1 {
		t.Fatalf("expected 1 trigger, got %d", len(page.Data))
	}

	got := page.Data[0]
	if len(got.ConditionIDs) != 1 || got.ConditionIDs[0] != 21 {
		t.Fatalf("unexpected condition IDs: %#v", got.ConditionIDs)
	}
	if got.ConditionCount != 1 || len(got.Conditions) != 1 {
		t.Fatalf("unexpected conditions enrichment: count=%d len=%d", got.ConditionCount, len(got.Conditions))
	}
	if got.SpaceID != 99 {
		t.Fatalf("unexpected space id: %d", got.SpaceID)
	}
	if len(got.SolutionIDs) != 1 || got.SolutionIDs[0] != 31 {
		t.Fatalf("unexpected solution IDs: %#v", got.SolutionIDs)
	}
	if got.SolutionCount != 1 || len(got.Solutions) != 1 {
		t.Fatalf("unexpected solutions enrichment: count=%d len=%d", got.SolutionCount, len(got.Solutions))
	}
	if got.LatestTaskID == nil || *got.LatestTaskID != 61 {
		t.Fatalf("unexpected latest task id: %#v", got.LatestTaskID)
	}
	if got.LatestTaskCreated == nil {
		t.Fatalf("expected latest task created to be set")
	}
	if got.LatestTaskResult == nil || *got.LatestTaskResult != constant.ScanTaskResultPass {
		t.Fatalf("unexpected latest task result: %#v", got.LatestTaskResult)
	}
	if got.LatestTaskStatus == nil || *got.LatestTaskStatus != constant.ScanTaskStatusSuccess {
		t.Fatalf("unexpected latest task status: %#v", got.LatestTaskStatus)
	}
}

func TestScanTriggerStoreGetByRepoPlanAndIDEnrichment(t *testing.T) {
	t.Parallel()

	db := newScanStoreTestDB(t)
	store := NewScanTriggerStore(db)
	ctx := context.Background()
	now := datetime.NowLocalTime()
	result := constant.ScanTaskResultPass

	repo := repoStore.Repository{ID: 202, SpaceID: 909, Identifier: "repo-202", Deleted: 0}
	if err := db.WithContext(ctx).Create(&repo).Error; err != nil {
		t.Fatalf("seed repo failed: %v", err)
	}

	trigger := ScanPlanTrigger{ID: 211, PlanID: 201, RepoID: 202, Name: "trigger-get", Type: string(scanModel.ScanTriggerTypeEvent), ScanType: string(constant.ScanTypeIncremental), Actions: string(constant.TriggerActionBranchUpdated), CreatedBy: "tester", CreatedDate: now, EditedBy: "tester", EditedDate: now}
	if err := db.WithContext(ctx).Create(&trigger).Error; err != nil {
		t.Fatalf("seed trigger failed: %v", err)
	}
	condition := ScanPlanCondition{ID: 221, RepoID: 202, PlanID: 201, TriggerID: 211, Priority: constant.ScanRulePriorityLow, Type: constant.ScanRuleTypeDefect, Unit: constant.ScanRuleUnitCount, Threshold: 1, CreatedBy: "tester", CreatedDate: now, EditedBy: "tester", EditedDate: now}
	if err := db.WithContext(ctx).Create(&condition).Error; err != nil {
		t.Fatalf("seed condition failed: %v", err)
	}
	solution := SolutionRules{ID: 231, Name: "solution-get", Desc: "desc", Tag: "tag", Status: constant.ScanRuleStatusEnabled, CreatedBy: "tester", CreatedDate: now, EditedBy: "tester", EditedDate: now}
	if err := db.WithContext(ctx).Create(&solution).Error; err != nil {
		t.Fatalf("seed solution failed: %v", err)
	}
	triggerSolution := ScanTriggerSolution{ID: 241, TriggerID: 211, SolutionID: 231, CreatedBy: "tester", CreatedDate: now, EditedBy: "tester", EditedDate: now}
	if err := db.WithContext(ctx).Create(&triggerSolution).Error; err != nil {
		t.Fatalf("seed trigger solution failed: %v", err)
	}
	execution := execStore.Execution{ID: 251, PipelineID: 1, Parent: 0, Trigger: "manual", Commit: "abc", Ref: "refs/heads/main", Params: "{}", Status: string(constant.CIStatusSuccess), CreatedBy: "tester", CreatedDate: now, EditedDate: now}
	if err := db.WithContext(ctx).Create(&execution).Error; err != nil {
		t.Fatalf("seed execution failed: %v", err)
	}
	task := ScanTask{ID: 261, PlanID: 201, ExecutionID: 251, RepoNumber: 202, PlanNumber: 201, TriggerID: ptrUint64(211), Status: constant.ScanTaskStatusSuccess, Result: &result, CreatedBy: "tester", CreatedDate: now, EditedDate: now, ScanType: constant.ScanTypeIncremental}
	if err := db.WithContext(ctx).Create(&task).Error; err != nil {
		t.Fatalf("seed task failed: %v", err)
	}

	got, err := store.GetByRepoPlanAndID(ctx, 202, 201, 211)
	if err != nil {
		t.Fatalf("GetByRepoPlanAndID returned error: %v", err)
	}
	if got.LatestTaskID != nil {
		t.Fatalf("expected detail latest task id to be nil, got %#v", got.LatestTaskID)
	}
	if got.LatestTaskStatus != nil {
		t.Fatalf("expected detail latest task status to be nil, got %#v", got.LatestTaskStatus)
	}
	if got.ConditionCount != 1 || got.SolutionCount != 1 {
		t.Fatalf("unexpected detail enrichment counts: condition=%d solution=%d", got.ConditionCount, got.SolutionCount)
	}
	if got.SpaceID != 909 {
		t.Fatalf("unexpected detail space id: %d", got.SpaceID)
	}
}

func TestScanPlanStoreListEnrichment(t *testing.T) {
	t.Parallel()

	db := newScanStoreTestDB(t)
	store := NewScanPlanStore(db)
	ctx := context.Background()
	now := datetime.NowLocalTime()
	result := constant.ScanTaskResultPass

	space := spaceStore.Space{ID: 88, Name: "space-88", Permission: spaceStore.AclOpen, CreatedBy: "tester", EditedBy: "tester", Deleted: 0}
	if err := db.WithContext(ctx).Create(&space).Error; err != nil {
		t.Fatalf("seed space failed: %v", err)
	}
	repo := repoStore.Repository{ID: 2, SpaceID: 88, Identifier: "repo-2", Deleted: 0}
	if err := db.WithContext(ctx).Create(&repo).Error; err != nil {
		t.Fatalf("seed repo failed: %v", err)
	}

	plan := ScanPlan{ID: 101, RepoID: 2, Name: "plan-1", ScanType: string(constant.ScanTypeIncremental), Branches: "[]", Files: "[]", CreatedBy: "tester", CreatedDate: now, EditedBy: "tester", EditedDate: now}
	if err := db.WithContext(ctx).Create(&plan).Error; err != nil {
		t.Fatalf("seed plan failed: %v", err)
	}
	solution := SolutionRules{ID: 131, Name: "solution-1", Desc: "desc", Tag: "tag", Status: constant.ScanRuleStatusEnabled, CreatedBy: "tester", CreatedDate: now, EditedBy: "tester", EditedDate: now}
	if err := db.WithContext(ctx).Create(&solution).Error; err != nil {
		t.Fatalf("seed solution failed: %v", err)
	}
	binding := ScanPlanSolution{ID: 141, PlanID: 101, SolutionID: 131, CreatedBy: "tester", CreatedDate: now, EditedBy: "tester", EditedDate: now}
	if err := db.WithContext(ctx).Create(&binding).Error; err != nil {
		t.Fatalf("seed plan solution failed: %v", err)
	}
	execution := execStore.Execution{ID: 151, PipelineID: 1, Parent: 0, Trigger: "manual", Commit: "abc", Ref: "refs/heads/main", Params: "{}", Status: string(constant.CIStatusFailure), CreatedBy: "tester", CreatedDate: now, EditedDate: now}
	if err := db.WithContext(ctx).Create(&execution).Error; err != nil {
		t.Fatalf("seed execution failed: %v", err)
	}
	task := ScanTask{ID: 161, PlanID: 101, ExecutionID: 151, RepoNumber: 2, PlanNumber: 101, Status: constant.ScanTaskStatusFailed, Result: &result, CreatedBy: "tester", CreatedDate: now, EditedDate: now, ScanType: constant.ScanTypeIncremental}
	if err := db.WithContext(ctx).Create(&task).Error; err != nil {
		t.Fatalf("seed task failed: %v", err)
	}

	page, err := store.List(ctx, 2, scanModel.ScanPlanReqFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Data) != 1 {
		t.Fatalf("expected 1 plan, got %d", len(page.Data))
	}
	got := page.Data[0]
	if len(got.SolutionIDs) != 1 || got.SolutionIDs[0] != 131 {
		t.Fatalf("unexpected solution IDs: %#v", got.SolutionIDs)
	}
	if got.RepoName != "repo-2" || got.SpaceID != 88 || got.SpaceName != "space-88" {
		t.Fatalf("unexpected repo/space enrichment: repo=%s spaceID=%d space=%s", got.RepoName, got.SpaceID, got.SpaceName)
	}
	if got.SolutionCount != 1 {
		t.Fatalf("unexpected solution count: %d", got.SolutionCount)
	}
	if got.LatestTaskID == nil || *got.LatestTaskID != 161 {
		t.Fatalf("unexpected latest task id: %#v", got.LatestTaskID)
	}
	if got.LatestTaskCreated == nil {
		t.Fatalf("expected latest task created to be set")
	}
	if got.LatestTaskResult == nil || *got.LatestTaskResult != constant.ScanTaskResultPass {
		t.Fatalf("unexpected latest task result: %#v", got.LatestTaskResult)
	}
	if got.LatestTaskStatus == nil || *got.LatestTaskStatus != constant.ScanTaskStatusFailed {
		t.Fatalf("unexpected latest task status: %#v", got.LatestTaskStatus)
	}
}

func ptrUint64(v uint64) *uint64 {
	return &v
}
