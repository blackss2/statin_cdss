package main

import ()

func Subject_AECount(SubjectId string) (int, error) {
	if cp, has := gAECountHash[SubjectId]; has {
		return cp, nil
	} else {
		return Update_AECount(SubjectId)
	}
}

func Update_AECount(SubjectId string) (int, error) {
	forms, err := FormWithCache(gAPI.FormTable, gConfig.StudyId)
	if err != nil {
		return 0, err
	}
	groupHash := make(map[string]*Group)
	itemHash := make(map[string]*Item)
	for _, form := range forms {
		if form.Id == "f-1" {
			addFormMeta(form, groupHash, itemHash)
		}
	}

	stacks, err := gAPI.StackTable.ListByFormIds([]string{"f-2", "f-3"}) //TEMP
	if err != nil {
		return 0, err
	}

	StackIds := make([]string, 0, len(stacks))
	for _, v := range stacks {
		if v.SubjectId == SubjectId {
			StackIds = append(StackIds, v.Id)
		}
	}
	visits, err := gAPI.VisitTable.ListByStackIds(StackIds)
	if err != nil {
		return 0, err
	}

	Ids := make([]string, 0, len(visits))
	visitHash := make(map[string]*Visit)
	for _, v := range visits {
		Ids = append(Ids, v.Id)
		visitHash[v.Id] = v
	}
	dataList, err := gAPI.DataTable.ListByVisitIds(Ids)
	if err != nil {
		return 0, err
	}

	countHash := make(map[int64]bool)
	for _, d := range dataList {
		if len(d.Value) > 0 || len(d.CodeId) > 0 {
			countHash[d.Rowindex] = true
		}
	}

	count := len(countHash)
	gAECountHash[SubjectId] = count
	return count, nil
}

var gAECountHash = make(map[string]int)
