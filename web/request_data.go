package web

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/gorilla/sessions"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/translation"
	dip "github.com/zond/godip/common"
)

type RequestData struct {
	response     http.ResponseWriter
	request      *http.Request
	session      *sessions.Session
	translations map[string]string
	web          *Web
}

func (self *Web) GetRequestData(w http.ResponseWriter, r *http.Request) (result RequestData) {
	result = RequestData{
		response:     w,
		request:      r,
		web:          self,
		translations: translation.GetTranslations(common.GetLanguage(r)),
	}
	result.session, _ = self.sessionStore.Get(r, SessionName)
	return
}

func (self RequestData) Close() {
	self.session.Save(self.request, self.response)
}

func (self RequestData) Appcache() bool {
	return self.web.appcache
}

func (self RequestData) AllocationMethodMap() string {
	result := map[string]common.AllocationMethod{}
	for _, meth := range common.AllocationMethodMap {
		meth.Translation = self.I(meth.Name)
		result[meth.Id] = meth
	}
	return common.Prettify(result)
}

func (self RequestData) SecrecyTypesMap() string {
	return common.Prettify(map[string]string{
		"SecretEmail":    self.I("Secret email"),
		"SecretNickname": self.I("Secret nickname"),
		"SecretNation":   self.I("Secret nation"),
	})
}

func (self RequestData) AllocationMethods() string {
	result := sort.StringSlice{}
	for _, meth := range common.AllocationMethods {
		result = append(result, meth.Id)
	}
	sort.Sort(result)
	return common.Prettify(result)
}

func (self RequestData) DefaultAllocationMethod() string {
	return common.RandomString
}

func (self RequestData) DefaultVariant() string {
	return common.ClassicalString
}

func (self RequestData) Variants() string {
	result := sort.StringSlice{}
	for _, variant := range common.Variants {
		result = append(result, variant.Id)
	}
	sort.Sort(result)
	return common.Prettify(result)
}

func (self RequestData) VariantMap() string {
	result := map[string]common.Variant{}
	for _, variant := range common.Variants {
		variant.Translation = self.I(variant.Name)
		result[variant.Id] = variant
	}
	return common.Prettify(result)
}

func (self RequestData) ChatFlagMap() string {
	return common.Prettify(map[string]int{
		"ChatPrivate":    common.ChatPrivate,
		"ChatGroup":      common.ChatGroup,
		"ChatConference": common.ChatConference,
	})
}

func (self RequestData) VariantColorizableProvincesMap() string {
	result := map[string][]dip.Province{}
	for _, variant := range common.Variants {
		supers := map[dip.Province]bool{}
		for _, prov := range variant.Graph.Provinces() {
			supers[prov.Super()] = true
		}
		for prov, _ := range supers {
			result[variant.Id] = append(result[variant.Id], prov)
		}
	}
	return common.Prettify(result)
}

func (self RequestData) VariantMainProvincesMap() string {
	result := map[string][]dip.Province{}
	for _, variant := range common.Variants {
		result[variant.Id] = []dip.Province{}
		for _, prov := range variant.Graph.Provinces() {
			if prov.Super() == prov {
				result[variant.Id] = append(result[variant.Id], prov)
			}
		}
	}
	return common.Prettify(result)
}

func (self RequestData) ChatFlagOptions() (result []common.ChatFlagOption) {
	for _, option := range common.ChatFlagOptions {
		result = append(result, common.ChatFlagOption{
			Id:          option.Id,
			Translation: self.I(option.Name),
		})
	}
	return
}

func (self RequestData) Authenticated() bool {
	return true
}

func (self RequestData) Abs(path string) string {
	return url.QueryEscape(fmt.Sprintf("http://%v%v", self.request.Host, path))
}

func (self RequestData) I(phrase string, args ...string) string {
	pattern, ok := self.translations[phrase]
	if !ok {
		panic(fmt.Errorf("Found no translation for %v", phrase))
	}
	if len(args) > 0 {
		return fmt.Sprintf(pattern, args)
	}
	return pattern
}

func (self RequestData) LogLevel() int {
	return self.web.logLevel
}

func (self RequestData) GameState(s string) common.GameState {
	switch s {
	case "Created":
		return common.GameStateCreated
	case "Started":
		return common.GameStateStarted
	case "Ended":
		return common.GameStateEnded
	}
	panic(fmt.Errorf("Unknown game state %v", s))
}

func (self RequestData) SecretFlagMap() string {
	return common.Prettify(map[string]int{
		"BeforeGame": common.SecretBeforeGame,
		"DuringGame": common.SecretDuringGame,
		"AfterGame":  common.SecretAfterGame,
	})
}

func (self RequestData) SecretFlag(s string) common.SecretFlag {
	switch s {
	case "BeforeGame":
		return common.SecretBeforeGame
	case "DuringGame":
		return common.SecretDuringGame
	case "AfterGame":
		return common.SecretAfterGame
	}
	panic(fmt.Errorf("Unknown secret flag %v", s))
}

func (self RequestData) ChatFlag(s string) string {
	var rval common.ChatFlag
	switch s {
	case "Private":
		rval = common.ChatPrivate
	case "Group":
		rval = common.ChatGroup
	case "Conference":
		rval = common.ChatConference
	}
	return fmt.Sprint(rval)
}

var version = time.Now()

func (self RequestData) Version() string {
	return fmt.Sprintf("%v", version.UnixNano())
}

func (self RequestData) SVG(p string) string {
	b := new(bytes.Buffer)
	if err := self.web.svgTemplates.ExecuteTemplate(b, p, self); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
	return string(b.Bytes())
}
