package main

import (
	"fmt"
)

type Compliance struct {
	Day               string `json:"day"`
	HasVisit          bool   `json:"has_visit"`
	LocalCount        int    `json:"localcount"`
	LocalItemCount    int    `json:"localitemcount"`
	SysmeticCount     int    `json:"sysmeticcount"`
	SysmeticItemCount int    `json:"sysmeticitemcount"`
	VitalCount        int    `json:"vitalcount"`
	VitalItemCount    int    `json:"vitalitemcount"`
	TotalCount        int    `json:"totalcount"`
	TotalItemCount    int    `json:"totalitemcount"`
}

type DAO_Compliance struct {
	Day      string `json:"day"`
	Local    int    `json:"local"`
	Sysmetic int    `json:"sysmetic"`
	Vital    int    `json:"vital"`
	Total    int    `json:"total"`
}

func Subject_Compliance(SubjectId string) ([]*Compliance, error) {
	if cp, has := gComplianceHash[SubjectId]; has {
		return cp, nil
	} else {
		return Update_Compliance(SubjectId)
	}
}

func Update_Compliance(SubjectId string) ([]*Compliance, error) {
	compliances := make([]*Compliance, 0)

	forms, err := FormWithCache(gAPI.FormTable, gConfig.StudyId)
	if err != nil {
		return nil, err
	}
	groupHash := make(map[string]*Group)
	itemHash := make(map[string]*Item)
	for _, form := range forms {
		if form.Id == "f-1" {
			addFormMeta(form, groupHash, itemHash)
		}
	}

	stack, err := gAPI.StackTable.Stack(SubjectId, "f-1") //TEMP
	if err != nil {
		if err != ErrNotExist {
			return nil, err
		}
	}

	DAY_COUNT := 7

	cpHash := make(map[string]*Compliance)
	for i := 0; i < DAY_COUNT; i++ {
		p := fmt.Sprintf("%d", i+1)
		cp := &Compliance{
			Day: p,
		}
		cpHash[p] = cp
		compliances = append(compliances, cp)
	}
	countHash := make(map[string]map[string]int)
	if stack != nil {
		visits, err := gAPI.VisitTable.List(stack.Id)
		if err != nil {
			return nil, err
		}

		Ids := make([]string, 0, len(visits))
		visitHash := make(map[string]*Visit)
		for _, v := range visits {
			Ids = append(Ids, v.Id)
			visitHash[v.Id] = v
			cpHash[v.Position].HasVisit = true
		}
		if len(Ids) > 0 {
			dataList, err := gAPI.DataTable.ListByVisitIds(Ids)
			if err != nil {
				return nil, err
			}
			for _, d := range dataList {
				if item, has := itemHash[d.ItemId]; has {
					if item.Type == "radio" {
						if visit, has := visitHash[d.VisitId]; has {
							hash, has := countHash[visit.Position]
							if !has {
								hash = make(map[string]int)
								countHash[visit.Position] = hash
							}
							hash[item.GroupId]++
						}
					}
				}
			}
		}
	}

	baseHash := make(map[string]int, DAY_COUNT)
	sumHash := make(map[string]int, DAY_COUNT)
	for i := 0; i < DAY_COUNT; i++ {
		p := fmt.Sprintf("%d", i+1)
		for _, v := range itemHash {
			if v.Type == "radio" {
				switch v.GroupId {
				case "g-2":
					cpHash[p].LocalItemCount++
				case "g-3":
					cpHash[p].SysmeticItemCount++
				case "g-7":
					cpHash[p].VitalItemCount++
				}
			}
		}
		baseHash[p] = cpHash[p].LocalItemCount + cpHash[p].SysmeticItemCount + cpHash[p].VitalItemCount
	}
	for p, hash := range countHash {
		for groupid, v := range hash {
			switch groupid {
			case "g-2":
				cpHash[p].LocalCount = v
			case "g-3":
				cpHash[p].SysmeticCount = v
			case "g-7":
				cpHash[p].VitalCount = v
			}
		}
		sumHash[p] = cpHash[p].LocalCount + cpHash[p].SysmeticCount + cpHash[p].VitalCount
	}
	for i := 0; i < DAY_COUNT; i++ {
		p := fmt.Sprintf("%d", i+1)
		cpHash[p].TotalCount = sumHash[p]
		cpHash[p].TotalItemCount = baseHash[p]
	}

	gComplianceHash[SubjectId] = compliances
	return compliances, nil
}

var gComplianceHash = make(map[string][]*Compliance)
