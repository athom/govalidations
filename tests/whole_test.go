package tests

import (
	"fmt"
	"github.com/sunfmin/govalidations"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

type User struct {
	Username  string
	FirstName string
	LastName  string
	Email     string
	Bio       string
	Age       int
}

var BlackListMessage = "You are in blacklist, lol :D"
var WhiteListMessage = "You are in whitelist, lol :D"

func MessageSwitcherGateKeeper() (gk *govalidations.GateKeeper) {
	gk = govalidations.NewGateKeeper()
	gk.Add(govalidations.MessageSwitcher(func(object interface{}) string {
		name := object.(*User).Username
		if name == "Kioshi" {
			return BlackListMessage
		}
		if name == "Roku" {
			return WhiteListMessage
		}
		return ""
	}, "Username"))
	return
}

func UserGateKeeper() (gk *govalidations.GateKeeper) {
	gk = govalidations.NewGateKeeper()

	gk.Add(govalidations.Regexp(func(object interface{}) interface{} {
		return object.(*User).Email
	}, regexp.MustCompile(`^([^@\s]+)@((?:[-a-z0-9]+\.)+[a-z]{2,})$`), "Email", "Must be a valid email"))

	gk.Add(govalidations.Presence(func(object interface{}) interface{} {
		return object.(*User).Username
	}, "Username", "Username can not be blank"))

	gk.Add(govalidations.Limitation(func(object interface{}) interface{} {
		return object.(*User).Username
	}, 0, 10, "Username", "Username can not be too long"))

	gk.Add(govalidations.Prohibition(func(object interface{}) interface{} {
		return object.(*User).Username
	}, 10, 20, "Username", "Username must less than 10 or more than 20"))

	gk.Add(govalidations.AvoidScriptTag(func(object interface{}) interface{} {
		return object.(*User).Username
	}, "Username", "Username can contains html script tag"))

	gk.Add(govalidations.Custom(func(object interface{}) bool {
		age := object.(*User).Age
		if age < 18 {
			return false
		}
		return true
	}, "Age", "You must be a grown man"))

	return
}

func theMux() (sm *http.ServeMux) {
	sm = http.NewServeMux()

	tpl := template.Must(template.ParseGlob("validate.html"))

	gk := UserGateKeeper()

	sm.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		u := &User{
			Username: "",
			Email:    "fake",
		}

		vd := gk.Validate(u)
		if vd.HasError() {
			tpl.Execute(w, vd)
			return
		}

		fmt.Fprintln(w, "Yeah!")
	})

	return
}

func aMux() (sm *http.ServeMux) {
	sm = http.NewServeMux()

	tpl := template.Must(template.ParseGlob("validate.html"))

	gk := UserGateKeeper()

	sm.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		u := &User{
			Username: "i like to move it move it",
			Email:    "kiss@therain.com",
		}

		vd := gk.Validate(u)
		if vd.HasError() {
			tpl.Execute(w, vd)
			return
		}

		fmt.Fprintln(w, "Yeah!")
	})

	return
}

func rokuMux() (sm *http.ServeMux) {
	sm = http.NewServeMux()
	tpl := template.Must(template.ParseGlob("validate.html"))
	gk := MessageSwitcherGateKeeper()
	sm.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		u := &User{
			Username: "Roku",
			Email:    "roku@avatar.com",
		}
		vd := gk.Validate(u)
		if vd.HasError() {
			tpl.Execute(w, vd)
			return
		}
		fmt.Fprintln(w, "Yeah!")
	})
	return
}

func kioshiMux() (sm *http.ServeMux) {
	sm = http.NewServeMux()
	tpl := template.Must(template.ParseGlob("validate.html"))
	gk := MessageSwitcherGateKeeper()
	sm.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		u := &User{
			Username: "Kioshi",
			Email:    "kioshi@avatar.com",
		}
		vd := gk.Validate(u)
		if vd.HasError() {
			tpl.Execute(w, vd)
			return
		}
		fmt.Fprintln(w, "Yeah!")
	})
	return
}

func avoidScriptTagMux() (sm *http.ServeMux) {
	sm = http.NewServeMux()

	tpl := template.Must(template.ParseGlob("validate.html"))

	gk := UserGateKeeper()

	sm.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		u := &User{
			Username: `<script>alert("x")<\script>`,
			Email:    "kiss@therain.com",
		}

		vd := gk.Validate(u)
		if vd.HasError() {
			tpl.Execute(w, vd)
			return
		}

		fmt.Fprintln(w, "Yeah!")
	})

	return
}

func prohibitionMux() (sm *http.ServeMux) {
	sm = http.NewServeMux()

	tpl := template.Must(template.ParseGlob("validate.html"))

	gk := UserGateKeeper()

	sm.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		u := &User{
			Username: "12345678901112",
			Email:    "kiss@therain.com",
		}

		vd := gk.Validate(u)
		if vd.HasError() {
			tpl.Execute(w, vd)
			return
		}

		fmt.Fprintln(w, "Yeah!")
	})

	return
}

func TestRenderErrors(t *testing.T) {
	ts := httptest.NewServer(theMux())
	defer ts.Close()

	r, _ := http.Get(ts.URL + "/validate")

	b, _ := ioutil.ReadAll(r.Body)
	body := string(b)
	if !strings.Contains(body, "Must be a valid email") {
		t.Error(body)
	}
	if !strings.Contains(body, "You must be a grown man") {
		t.Error(body)
	}
	if !strings.Contains(body, "Username can not be blank") {
		t.Error(body)
	}
}

func TestRenderLimitationErrors(t *testing.T) {
	ts := httptest.NewServer(aMux())
	defer ts.Close()

	r, _ := http.Get(ts.URL + "/validate")

	b, _ := ioutil.ReadAll(r.Body)
	body := string(b)
	if !strings.Contains(body, "Username can not be too long") {
		t.Error(body)
	}
}

func TestRenderAvoidScriptTagErrors(t *testing.T) {
	ts := httptest.NewServer(avoidScriptTagMux())
	defer ts.Close()

	r, _ := http.Get(ts.URL + "/validate")

	b, _ := ioutil.ReadAll(r.Body)
	body := string(b)
	if !strings.Contains(body, "Username can contains html script tag") {
		t.Error(body)
	}
}

func TestRenderMessageSwitcher(t *testing.T) {
	ts := httptest.NewServer(kioshiMux())
	defer ts.Close()

	r, _ := http.Get(ts.URL + "/validate")

	b, _ := ioutil.ReadAll(r.Body)
	body := string(b)
	if !strings.Contains(body, BlackListMessage) {
		t.Error(body)
	}
	if strings.Contains(body, WhiteListMessage) {
		t.Error(body)
	}

	ts = httptest.NewServer(rokuMux())
	defer ts.Close()

	r, _ = http.Get(ts.URL + "/validate")

	b, _ = ioutil.ReadAll(r.Body)
	body = string(b)
	if !strings.Contains(body, WhiteListMessage) {
		t.Error(body)
	}
	if strings.Contains(body, BlackListMessage) {
		t.Error(body)
	}
}

func TestRenderProhibitionErrors(t *testing.T) {
	ts := httptest.NewServer(prohibitionMux())
	defer ts.Close()

	r, _ := http.Get(ts.URL + "/validate")

	b, _ := ioutil.ReadAll(r.Body)
	body := string(b)
	if !strings.Contains(body, "Username must less than 10 or more than 20") {
		t.Error(body)
	}
}
