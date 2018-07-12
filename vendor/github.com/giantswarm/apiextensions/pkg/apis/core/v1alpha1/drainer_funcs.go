package v1alpha1

func (s DrainerConfigStatus) HasDrainedCondition() bool {
	return hasDrainerConfigCondition(s.Conditions, DrainerConfigStatusStatusTrue, DrainerConfigStatusTypeDrained)
}

func (s DrainerConfigStatus) HasTimeoutCondition() bool {
	return hasDrainerConfigCondition(s.Conditions, DrainerConfigStatusStatusTrue, DrainerConfigStatusTypeTimeout)
}

func (s DrainerConfigStatus) NewDrainedCondition() DrainerConfigStatusCondition {
	return DrainerConfigStatusCondition{
		Status: DrainerConfigStatusStatusTrue,
		Type:   DrainerConfigStatusTypeDrained,
	}
}

func (s DrainerConfigStatus) NewTimeoutCondition() DrainerConfigStatusCondition {
	return DrainerConfigStatusCondition{
		Status: DrainerConfigStatusStatusTrue,
		Type:   DrainerConfigStatusTypeDrained,
	}
}

func hasDrainerConfigCondition(conditions []DrainerConfigStatusCondition, search string, status string) bool {
	for _, c := range conditions {
		if c.Status == search && c.Type == status {
			return true
		}
	}

	return false
}
