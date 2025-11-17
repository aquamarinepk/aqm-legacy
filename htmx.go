package aqm

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
)

// HTMX request headers
const (
	HXRequest      = "HX-Request"
	HXTrigger      = "HX-Trigger"
	HXTriggerName  = "HX-Trigger-Name"
	HXTarget       = "HX-Target"
	HXCurrentURL   = "HX-Current-URL"
	HXPrompt       = "HX-Prompt"
	HXBoosted      = "HX-Boosted"
	HXHistoryState = "HX-History-Restore-Request"
)

// HTMX response headers
const (
	HXLocation    = "HX-Location"
	HXPushURL     = "HX-Push-Url"
	HXRedirect    = "HX-Redirect"
	HXRefresh     = "HX-Refresh"
	HXReplaceURL  = "HX-Replace-Url"
	HXReswap      = "HX-Reswap"
	HXRetarget    = "HX-Retarget"
	HXReselect    = "HX-Reselect"
	HXTriggerResp = "HX-Trigger"
)

// SwapMode defines how HTMX swaps content into the DOM.
type SwapMode string

const (
	SwapInnerHTML   SwapMode = "innerHTML"
	SwapOuterHTML   SwapMode = "outerHTML"
	SwapBeforeBegin SwapMode = "beforebegin"
	SwapAfterBegin  SwapMode = "afterbegin"
	SwapBeforeEnd   SwapMode = "beforeend"
	SwapAfterEnd    SwapMode = "afterend"
	SwapDelete      SwapMode = "delete"
	SwapNone        SwapMode = "none"
)

// IsHTMX returns true if the request was initiated by HTMX.
func IsHTMX(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get(HXRequest), "true")
}

// IsBoosted returns true if the request was boosted by HTMX.
func IsBoosted(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get(HXBoosted), "true")
}

// IsHistoryRestore returns true if the request is for history restoration.
func IsHistoryRestore(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get(HXHistoryState), "true")
}

// GetHTMXTrigger returns the ID of the element that triggered the request.
func GetHTMXTrigger(r *http.Request) string {
	return r.Header.Get(HXTrigger)
}

// GetHTMXTriggerName returns the name of the element that triggered the request.
func GetHTMXTriggerName(r *http.Request) string {
	return r.Header.Get(HXTriggerName)
}

// GetHTMXTarget returns the ID of the target element.
func GetHTMXTarget(r *http.Request) string {
	return r.Header.Get(HXTarget)
}

// GetHTMXCurrentURL returns the current URL of the browser.
func GetHTMXCurrentURL(r *http.Request) string {
	return r.Header.Get(HXCurrentURL)
}

// GetHTMXPrompt returns the user response to hx-prompt if it was used.
func GetHTMXPrompt(r *http.Request) string {
	return r.Header.Get(HXPrompt)
}

// SetHTMXRedirect sets the HX-Redirect header to perform a client-side redirect.
func SetHTMXRedirect(w http.ResponseWriter, url string) {
	w.Header().Set(HXRedirect, url)
}

// SetHTMXRefresh sets the HX-Refresh header to do a full page refresh.
func SetHTMXRefresh(w http.ResponseWriter) {
	w.Header().Set(HXRefresh, "true")
}

// SetHTMXPushURL sets the HX-Push-Url header to push a new URL into the history.
func SetHTMXPushURL(w http.ResponseWriter, url string) {
	w.Header().Set(HXPushURL, url)
}

// SetHTMXReplaceURL sets the HX-Replace-Url header to replace the current URL.
func SetHTMXReplaceURL(w http.ResponseWriter, url string) {
	w.Header().Set(HXReplaceURL, url)
}

// SetHTMXRetarget changes the target of the content update to a different element.
func SetHTMXRetarget(w http.ResponseWriter, selector string) {
	w.Header().Set(HXRetarget, selector)
}

// SetHTMXReswap changes the swap mode of the response.
func SetHTMXReswap(w http.ResponseWriter, mode SwapMode) {
	w.Header().Set(HXReswap, string(mode))
}

// SetHTMXReselect allows you to select a subset of the response to swap.
func SetHTMXReselect(w http.ResponseWriter, selector string) {
	w.Header().Set(HXReselect, selector)
}

// TriggerEvent sends an HX-Trigger header to trigger client-side events.
// For simple event names, pass a string. For events with payloads, pass a map.
func TriggerEvent(w http.ResponseWriter, event interface{}) error {
	switch v := event.(type) {
	case string:
		w.Header().Set(HXTriggerResp, v)
		return nil
	case map[string]interface{}:
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		w.Header().Set(HXTriggerResp, string(data))
		return nil
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		w.Header().Set(HXTriggerResp, string(data))
		return nil
	}
}

// RedirectOrHeader performs an HTMX-aware redirect. If the request is from HTMX,
// it sets the HX-Redirect header. Otherwise, it performs a standard HTTP redirect.
func RedirectOrHeader(w http.ResponseWriter, r *http.Request, url string) {
	if IsHTMX(r) {
		SetHTMXRedirect(w, url)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// HTMX provides a fluent builder for HTMX attributes to be used in templates.
type HTMX struct {
	get       string
	post      string
	put       string
	patch     string
	delete    string
	target    string
	swap      SwapMode
	trigger   string
	pushURL   bool
	boost     bool
	confirm   string
	include   string
	vals      string
	select_   string
	indicator string
}

// H creates a new HTMX attribute builder.
func H() *HTMX {
	return &HTMX{}
}

// Get sets the hx-get attribute.
func (h *HTMX) Get(url string) *HTMX {
	h.get = url
	return h
}

// Post sets the hx-post attribute.
func (h *HTMX) Post(url string) *HTMX {
	h.post = url
	return h
}

// Put sets the hx-put attribute.
func (h *HTMX) Put(url string) *HTMX {
	h.put = url
	return h
}

// Patch sets the hx-patch attribute.
func (h *HTMX) Patch(url string) *HTMX {
	h.patch = url
	return h
}

// Delete sets the hx-delete attribute.
func (h *HTMX) Delete(url string) *HTMX {
	h.delete = url
	return h
}

// Target sets the hx-target attribute.
func (h *HTMX) Target(selector string) *HTMX {
	h.target = selector
	return h
}

// TargetID sets the hx-target attribute to a specific ID (adds # prefix).
func (h *HTMX) TargetID(id string) *HTMX {
	h.target = "#" + id
	return h
}

// Swap sets the hx-swap attribute.
func (h *HTMX) Swap(mode SwapMode) *HTMX {
	h.swap = mode
	return h
}

// Trigger sets the hx-trigger attribute.
func (h *HTMX) Trigger(event string) *HTMX {
	h.trigger = event
	return h
}

// PushURL enables hx-push-url.
func (h *HTMX) PushURL(enabled bool) *HTMX {
	h.pushURL = enabled
	return h
}

// Boost enables hx-boost.
func (h *HTMX) Boost() *HTMX {
	h.boost = true
	return h
}

// Confirm sets the hx-confirm attribute.
func (h *HTMX) Confirm(message string) *HTMX {
	h.confirm = message
	return h
}

// Include sets the hx-include attribute.
func (h *HTMX) Include(selector string) *HTMX {
	h.include = selector
	return h
}

// Vals sets the hx-vals attribute (JSON string).
func (h *HTMX) Vals(jsonStr string) *HTMX {
	h.vals = jsonStr
	return h
}

// Select sets the hx-select attribute.
func (h *HTMX) Select(selector string) *HTMX {
	h.select_ = selector
	return h
}

// Indicator sets the hx-indicator attribute.
func (h *HTMX) Indicator(selector string) *HTMX {
	h.indicator = selector
	return h
}

// Attrs returns the accumulated HTMX attributes as template.HTMLAttr for safe rendering.
func (h *HTMX) Attrs() template.HTMLAttr {
	var attrs []string

	if h.get != "" {
		attrs = append(attrs, `hx-get="`+template.HTMLEscapeString(h.get)+`"`)
	}
	if h.post != "" {
		attrs = append(attrs, `hx-post="`+template.HTMLEscapeString(h.post)+`"`)
	}
	if h.put != "" {
		attrs = append(attrs, `hx-put="`+template.HTMLEscapeString(h.put)+`"`)
	}
	if h.patch != "" {
		attrs = append(attrs, `hx-patch="`+template.HTMLEscapeString(h.patch)+`"`)
	}
	if h.delete != "" {
		attrs = append(attrs, `hx-delete="`+template.HTMLEscapeString(h.delete)+`"`)
	}
	if h.target != "" {
		attrs = append(attrs, `hx-target="`+template.HTMLEscapeString(h.target)+`"`)
	}
	if h.swap != "" {
		attrs = append(attrs, `hx-swap="`+template.HTMLEscapeString(string(h.swap))+`"`)
	}
	if h.trigger != "" {
		attrs = append(attrs, `hx-trigger="`+template.HTMLEscapeString(h.trigger)+`"`)
	}
	if h.pushURL {
		attrs = append(attrs, `hx-push-url="true"`)
	}
	if h.boost {
		attrs = append(attrs, `hx-boost="true"`)
	}
	if h.confirm != "" {
		attrs = append(attrs, `hx-confirm="`+template.HTMLEscapeString(h.confirm)+`"`)
	}
	if h.include != "" {
		attrs = append(attrs, `hx-include="`+template.HTMLEscapeString(h.include)+`"`)
	}
	if h.vals != "" {
		attrs = append(attrs, `hx-vals='`+h.vals+`'`)
	}
	if h.select_ != "" {
		attrs = append(attrs, `hx-select="`+template.HTMLEscapeString(h.select_)+`"`)
	}
	if h.indicator != "" {
		attrs = append(attrs, `hx-indicator="`+template.HTMLEscapeString(h.indicator)+`"`)
	}

	return template.HTMLAttr(strings.Join(attrs, " "))
}

// String returns the attributes as a plain string (for debugging).
func (h *HTMX) String() string {
	return string(h.Attrs())
}
