package buildcfg

type multiString struct {
	V []string
}

func (sm *multiString) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var multi []string
	err := unmarshal(&multi)
	if err != nil {
		var single string
		err := unmarshal(&single)
		if err != nil {
			return err
		}
		sm.V = make([]string, 1)
		sm.V[0] = single
	} else {
		sm.V = multi
	}
	return nil
}
