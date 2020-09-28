package ide

import (
	"syscall/js"

	"github.com/johnstarich/go-wasm/internal/interop"
)

type Tabber interface {
	Titles() <-chan string
}

type TabPane struct {
	lastTabID         int
	jsValue           js.Value
	tabButtonsParent  js.Value
	tabsParent        js.Value
	newTabListener    js.Func
	tabs              []*Tab
	currentTab        int
	makeDefaultTab    TabBuilder
	newTabOptions     TabOptions
	closedTabListener func(index int)
}

type TabOptions struct {
	NoFocus bool
}

type TabBuilder func(title, contents js.Value) Tabber

func NewTabPane(newTabOptions TabOptions, makeDefaultTab TabBuilder, closedTab func(index int)) *TabPane {
	elem := document.Call("createElement", "div")
	elem.Get("classList").Call("add", "pane")
	elem.Set("innerHTML", `
<nav class="tab-bar">
	<ul class="tab-buttons"></ul>
	<button class="tab-new"></button>
</nav>
<div class="tabs"></div>
`)
	p := &TabPane{
		jsValue:           elem,
		tabButtonsParent:  elem.Call("querySelector", ".tab-buttons"),
		tabsParent:        elem.Call("querySelector", ".tabs"),
		tabs:              nil,
		currentTab:        -1,
		makeDefaultTab:    makeDefaultTab,
		newTabOptions:     newTabOptions,
		closedTabListener: closedTab,
	}
	p.newTabListener = js.FuncOf(func(js.Value, []js.Value) interface{} {
		p.NewTab(newTabOptions, p.makeDefaultTab)
		return nil
	})
	elem.Call("querySelector", ".tab-new").Call("addEventListener", "click", p.newTabListener)
	return p
}

func (p *TabPane) JSValue() js.Value {
	return p.jsValue
}

func (p *TabPane) NewDefaultTab(options TabOptions) Tabber {
	return p.NewTab(options, p.makeDefaultTab)
}

func (p *TabPane) NewTab(options TabOptions, makeTab TabBuilder) Tabber {
	contents := document.Call("createElement", "div")
	contents.Set("className", "tab")
	p.tabsParent.Call("appendChild", contents)

	tabItem := document.Call("createElement", "li")
	tabItem.Get("classList").Call("add", "tab-button")
	tabItem.Set("innerHTML", `
<span class="tab-title">New file</span>
<button class="tab-close" title="close"></button>
`)
	title := tabItem.Call("querySelector", ".tab-title")
	p.tabButtonsParent.Call("appendChild", tabItem)

	tabber := makeTab(title, contents)
	tab := newTab(p.lastTabID, tabItem, contents, title, tabber, p.focusID)
	p.lastTabID++
	p.tabs = append(p.tabs, tab)

	closeButton := tabItem.Call("querySelector", ".tab-close")
	closeButton.Call("addEventListener", "click", interop.SingleUseFunc(func(_ js.Value, args []js.Value) interface{} {
		event := args[0]
		event.Call("stopPropagation")
		p.closeTabID(tab.id)
		return nil
	}))

	if !options.NoFocus {
		p.focusID(tab.id)
	}
	return tabber
}

func (p *TabPane) Focus(index int) {
	if index >= 0 {
		p.focusID(p.tabs[index].id)
	}
}

func (p *TabPane) focusID(id int) {
	if p.currentTab >= 0 {
		p.tabs[p.currentTab].Unfocus()
	}
	for i, tab := range p.tabs {
		if tab.id == id {
			p.currentTab = i
			tab.Focus()
		}
	}
}

func (p *TabPane) Close() {
	p.newTabListener.Release()
}

func (p *TabPane) closeTabID(id int) {
	var tabIndex int
	var tab *Tab
	for i, t := range p.tabs {
		if t.id == id {
			tabIndex = i
			tab = t
			break
		}
	}
	if tab == nil {
		return
	}

	tab.Close()
	p.tabButtonsParent.Get("children").Index(tabIndex).Call("remove")
	p.tabsParent.Get("children").Index(tabIndex).Call("remove")

	var newTabs []*Tab
	newTabs = append(newTabs, p.tabs[:tabIndex]...)
	newTabs = append(newTabs, p.tabs[tabIndex+1:]...)
	p.tabs = newTabs
	if p.currentTab == len(p.tabs) {
		p.currentTab = len(p.tabs) - 1
	}
	p.Focus(p.currentTab)

	p.closedTabListener(tabIndex)
}
