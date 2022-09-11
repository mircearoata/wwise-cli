package product

func (f File) GetGroupValue(groupId string) string {
	for _, group := range f.Groups {
		if group.GroupID == groupId {
			return group.GroupValueID
		}
	}
	return ""
}

type GroupFilter struct {
	GroupID     string
	GroupValues []string
}

func (pvi ProductVersionInfo) FindFilesByGroups(groupsFilters []GroupFilter) []File {
	var files []File
	for _, file := range pvi.Files {
		ok := true
		for _, group := range groupsFilters {
			fileGroupValue := file.GetGroupValue(group.GroupID)
			hasGroupValue := false
			for _, groupValue := range group.GroupValues {
				if fileGroupValue == groupValue {
					hasGroupValue = true
					break
				}
			}
			if !hasGroupValue {
				ok = false
				break
			}
		}
		if ok {
			files = append(files, file)
		}
	}
	return files
}
