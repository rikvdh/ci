package buildcfg

import (
	"encoding/json"
	"io/ioutil"
)

type repoData struct {
	Alias      string
	Sourceline string
	KeyUrl     string
}

var repoList map[string]repoData

func translateRepo(orig string) (string, string) {
	if repoList == nil {
		str, err := ioutil.ReadFile("./travis-apt-whitelist.json")
		if err != nil {
			panic(err)
		}
		var tmp []repoData
		if err := json.Unmarshal(str, &tmp); err != nil {
			panic(err)
		}
		repoList = make(map[string]repoData)
		for _, a := range tmp {
			repoList[a.Alias] = a
		}
	}
	list, ok := repoList[orig]
	if ok {
		return list.Sourceline, list.KeyUrl
	}
	return "", ""
}
