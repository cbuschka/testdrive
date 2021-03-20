package processing

type SigintEvent struct {
}

func (event *SigintEvent) Type() string {
	return "sigint"
}

func (event *SigintEvent) String() string {
	return "sigint"
}

func (event *SigintEvent) Id() string {
	return ""
}
