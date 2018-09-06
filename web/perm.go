package web

import (
	"net/url"
	"os"
	"path"

	"github.com/goftp/server"
	"github.com/tango-contrib/renders"
)

func hasPerm(mode os.FileMode, idx int, rOrW string) bool {
	return string(mode.String()[idx]) == rOrW
}

type PermBaseAction struct {
	BaseAuthAction
}

func (c *PermBaseAction) Before() {
	c.curModule = PERM_MODULE
}

type PermAction struct {
	PermBaseAction
}

func (c *PermAction) Get() error {
	var err error
	var parent string
	p := c.Form("path")
	if p != "" {
		p, err = url.QueryUnescape(p)
		if err != nil {
			return err
		}
		parent = p
	} else {
		parent = "/"
		p = "/"
	}
	driver, err := Factory.NewDriver()
	if err != nil {
		return err
	}
	var pathinfos []server.FileInfo
	err = driver.ListDir(p, func(f server.FileInfo) error {
		if f.Name() != "." {
			pathinfos = append(pathinfos, f)
		}
		return nil
	})
	if err != nil {
		return err
	}

	var users []User
	err = DB.UserList(&users)
	if err != nil {
		return err
	}
	var groups []string
	err = DB.GroupList(&groups)
	if err != nil {
		return err
	}
	return c.Render("perm/list.html", renders.T{
		"parent":  parent,
		"path":    p,
		"infos":   pathinfos,
		"hasPerm": hasPerm,
		"users":   users,
		"groups":  groups,
		"subPath": func(subPath string) string {
			return url.QueryEscape(path.Join(parent, subPath))
		},
	})
}

type PermAddAction struct {
	PermBaseAction
}

func (p *PermAddAction) Get() error {
	return nil
}

type PermEditAction struct {
	PermBaseAction
}

func (p *PermEditAction) Get() error {
	return nil
}

type PermDelAction struct {
	PermBaseAction
}

func (p *PermDelAction) Get() error {
	return nil
}

type PermUpdateGroup struct {
	PermBaseAction
}

func (c *PermUpdateGroup) Get() {
	name := c.Form("name")
	newgroup := c.Form("newgroup")

	if name == "" || newgroup == "" {
		c.ServeJson(map[string]string{"status": "0", "error": "empty params"})
		return
	}

	name, err := url.QueryUnescape(name)
	if err != nil {
		c.ServeJson(map[string]string{"status": "0", "error": err.Error()})
		return
	}

	err = Perm.ChGroup(name, newgroup)
	if err != nil {
		c.ServeJson(map[string]string{"status": "0", "error": err.Error()})
		return
	}
	c.ServeJson(map[string]string{"status": "1"})
}

type PermUpdateOwner struct {
	PermBaseAction
}

func (c *PermUpdateOwner) Get() {
	name := c.Form("name")
	newowner := c.Form("newowner")

	if name == "" || newowner == "" {
		c.ServeJson(map[string]string{"status": "0", "error": "empty params"})
		return
	}

	name, err := url.QueryUnescape(name)
	if err != nil {
		c.ServeJson(map[string]string{"status": "0", "error": err.Error()})
		return
	}

	err = Perm.ChOwner(name, newowner)
	if err != nil {
		c.ServeJson(map[string]string{"status": "0", "error": err.Error()})
		return
	}
	c.ServeJson(map[string]string{"status": "1"})
}

type PermUpdatePerm struct {
	PermBaseAction
}

func (c *PermUpdatePerm) Get() {
	name := c.Form("name")
	typ := c.Form("typ")
	right := c.Form("right")
	has := c.Form("has")

	if name == "" || typ == "" || right == "" || has == "" {
		c.ServeJson(map[string]string{"status": "0", "error": "empty params"})
		return
	}

	name, err := url.QueryUnescape(name)
	if err != nil {
		c.ServeJson(map[string]string{"status": "0", "error": err.Error()})
		return
	}

	mode, err := Perm.GetMode(name)
	if err != nil {
		c.ServeJson(map[string]string{"status": "0", "error": err.Error()})
		return
	}

	var bs uint = 0
	if typ == "owner" {
		if right == "r" {
			bs = 8
		} else if right == "w" {
			bs = 7
		}
	} else if typ == "group" {
		if right == "r" {
			bs = 5
		} else if right == "w" {
			bs = 4
		}
	} else if typ == "other" {
		if right == "r" {
			bs = 2
		} else if right == "w" {
			bs = 1
		}
	}

	if has == "true" {
		mode = os.FileMode(uint32(mode) + uint32(1<<bs))
	} else {
		mode = os.FileMode(uint32(mode) - uint32(1<<bs))
	}

	err = Perm.ChMode(name, mode)
	if err != nil {
		c.ServeJson(map[string]string{"status": "0", "error": err.Error()})
		return
	}
	c.ServeJson(map[string]string{"status": "1"})
}
