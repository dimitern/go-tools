package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type APIDescription struct {
	Doc       string        `json:"doc"`
	Hash      string        `json:"hash"`
	Resources []APIResource `json:"resources"`
}

type APIResource struct {
	Anon *APIHandler `json:"anon,omitempty"`
	Name string      `json:"name"`
	Auth *APIHandler `json:"auth,omitempty"`
}

type apiResourcesByPath []APIResource

func (a apiResourcesByPath) Len() int      { return len(a) }
func (a apiResourcesByPath) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a apiResourcesByPath) Less(i, j int) bool {
	anon1, anon2 := a[i].Anon, a[j].Anon
	hasAnon := anon1 != nil && anon2 != nil
	auth1, auth2 := a[i].Auth, a[j].Auth
	hasAuth := auth1 != nil && auth2 != nil

	switch {
	case hasAnon && !hasAuth:
		return anon1.Path < anon2.Path

	case !hasAnon && hasAuth:
		return auth1.Path < auth2.Path

	case hasAnon && hasAuth:
		return anon1.Path < anon2.Path && auth1.Path < auth2.Path
	}
	return false
}

type APIHandler struct {
	Name    string      `json:"name"`
	Doc     string      `json:"doc"`
	URI     string      `json:"uri"`
	Actions []APIAction `json:"actions"`
	Params  []string    `json:"params"`
	Path    string      `json:"path"`
}

type ActionParam struct {
	Name       string
	Doc        string
	PythonType string
	GoType     reflect.Type
}

func deriveGoTypeFromPythonType(pythonType string) reflect.Type {
	type unknown struct{}
	type object struct{}

	pt := strings.ToLower(pythonType)
	switch {
	case strings.Contains(pt, "ip address"):
		return reflect.TypeOf(net.IP{})
	case strings.HasPrefix(pt, "unicode"), strings.HasPrefix(pt, "string"):
		return reflect.TypeOf("string")
	case strings.HasPrefix(pt, "list of unicode"), strings.HasPrefix(pt, "iterable"):
		return reflect.SliceOf(reflect.TypeOf("string"))
	case strings.Contains(pt, "unicode"): // e.g. base64-encoded unicode
		return reflect.TypeOf("string")
	case strings.HasPrefix(pt, "bool"):
		return reflect.TypeOf(false)
	case strings.HasPrefix(pt, "int"):
		return reflect.TypeOf(int(0))
	case strings.HasPrefix(pt, "float"):
		return reflect.TypeOf(0.5)
	case strings.Contains(pt, "object"):
		return reflect.TypeOf(object{})
	}
	return reflect.TypeOf(unknown{})
}

type APIAction struct {
	RawDoc string `json:"doc"`
	// Doc contains the parsed first line of RawDoc.
	Doc string `json:"-"`
	// DocParams contains the parsed param name to description.
	DocParams map[string]ActionParam `json:"-"`
	// DocReturns contains the parsed HTTP status code to description(s).
	DocReturns map[int][]string `json:"-"`

	Op      *string `json:"op,omitempty"`
	Restful bool    `json:"restful"`
	Method  string  `json:"method"`
	Name    string  `json:"name"`
}

var (
	rawDocParams            = regexp.MustCompile(`(?::param )(?P<param>[^: ]+)(?:: )(?P<doc>(\n|.)+.?)`)
	rawDocParamType         = regexp.MustCompile(`(?::type )(?P<param>[^:]+)(?:: )(?P<type>(\n|.)+)$`)
	rawDocReturns           = regexp.MustCompile(`(?P<first>[Rr])(?:eturns )(?P<code>\d{3})(?: )(?P<doc>(\n|[^\.])+\.)`)
	rawDocExtraSpaces       = regexp.MustCompile(` {2,}`)
	rawDocParamsExtraSpaces = regexp.MustCompile(`( {2,}|\n |\n)`)
)

func (a APIAction) parseRawDoc() (
	doc string,
	docParams map[string]ActionParam,
	docReturns map[int][]string,
	err error,
) {
	// Separate pure doc lines from params and returns.
	rawDoc := a.RawDoc
	rawDocParts := strings.Split(rawDoc, "\n\n")
	rawDoc = ""
	var rawParamsAndReturns string
	for _, part := range rawDocParts {
		if rawDocParams.MatchString(part) || rawDocReturns.MatchString(part) {
			rawParamsAndReturns += part + "\n\n"
			continue
		}
		rawDoc += part + "\n\n"
	}

	// Strip extra whitespace.
	rawDoc = rawDocExtraSpaces.ReplaceAllString(rawDoc, " ")
	rawDoc = strings.TrimSpace(rawDoc)
	rawParamsAndReturns = rawDocParamsExtraSpaces.ReplaceAllString(rawParamsAndReturns, " ")
	rawParamsAndReturns = strings.Replace(rawParamsAndReturns, "  ", " ", -1)

	// Parse and extract all parsable returns.
	lastParamsAndReturns := ""
	for rawDocReturns.MatchString(rawParamsAndReturns) && lastParamsAndReturns != rawParamsAndReturns {
		matches := rawDocReturns.FindStringSubmatch(rawParamsAndReturns)
		if docReturns == nil {
			docReturns = make(map[int][]string)
		}
		firstLetter, code, doc := matches[1], matches[2], matches[3]
		httpCode, err := strconv.Atoi(code)
		if err != nil {
			return "", nil, nil, fmt.Errorf("unexpected HTTP code %q: %v", code, err)
		}
		if existing, ok := docReturns[httpCode]; ok {
			docReturns[httpCode] = append(existing, doc)
		} else {
			docReturns[httpCode] = append([]string(nil), doc)
		}

		rawReturn := fmt.Sprintf("%seturns %d %s", firstLetter, httpCode, doc)
		lastParamsAndReturns = rawParamsAndReturns
		rawParamsAndReturns = strings.Replace(rawParamsAndReturns, rawReturn, "", -1)
	}

	// Unshift all parsable params and their docs into docParams.
	lastParamsAndReturns = ""
	for rawDocParams.MatchString(rawParamsAndReturns) && lastParamsAndReturns != rawParamsAndReturns {
		matches := rawDocParams.FindStringSubmatch(rawParamsAndReturns)
		if docParams == nil {
			docParams = make(map[string]ActionParam)
		}
		param, doc := matches[1], matches[2]
		nextParamIndex := strings.Index(doc, " :param ")
		if nextParamIndex != -1 {
			doc = doc[:nextParamIndex]
		}
		doc = strings.TrimSpace(doc)
		paramName := param
		pythonType := "<unspecified>"
		docWithoutType := doc
		if rawDocParamType.MatchString(doc) {
			matches = rawDocParamType.FindStringSubmatch(doc)
			paramName, pythonType = matches[1], matches[2]
			rawType := fmt.Sprintf(":type %s: %s", paramName, pythonType)
			docWithoutType = strings.Replace(doc, rawType, "", -1)
			docWithoutType = strings.TrimSpace(docWithoutType)
		}
		pythonType = strings.TrimSpace(pythonType)

		docParams[param] = ActionParam{
			Name:       param,
			Doc:        docWithoutType,
			PythonType: pythonType,
			GoType:     deriveGoTypeFromPythonType(pythonType),
		}
		rawParam := fmt.Sprintf(":param %s: %s", param, doc)
		lastParamsAndReturns = rawParamsAndReturns
		rawParamsAndReturns = strings.Replace(rawParamsAndReturns, rawParam, "", -1)
	}

	return rawDoc, docParams, docReturns, nil
}

// GetAPIDescription takes a MAAS API URL prefix (e.g.
// "http://10.10.19.2/MAAS/") and returns the parsed APIDescription and the
// indented raw JSON, or an error.
func GetAPIDescription(apiPrefix string) (*APIDescription, string, error) {
	urlPrefix, err := url.Parse(apiPrefix)
	if err != nil {
		return nil, "", fmt.Errorf("cannot parse URL prefix %q: %v", apiPrefix, err)
	}

	fullURL, err := urlPrefix.Parse("api/1.0/describe/")
	if err != nil {
		return nil, "", fmt.Errorf("cannot parse full URL %q: %v", urlPrefix.String()+"describe/", err)
	}

	response, err := http.Get(fullURL.String())
	if err != nil {
		return nil, "", fmt.Errorf("cannot get API description at %q: %v", fullURL.String(), err)
	}

	defer response.Body.Close()
	bodyData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, "", fmt.Errorf("cannot read response body: %v", err)
	}

	var apiDescription APIDescription
	if err := json.Unmarshal(bodyData, &apiDescription); err != nil {
		return nil, "", fmt.Errorf("cannot unmarshal response JSON: %v", err)
	}
	var rawJSON map[string]interface{}
	if err := json.Unmarshal(bodyData, &rawJSON); err != nil {
		return nil, "", fmt.Errorf("cannot unmarshal response raw JSON: %v", err)
	}

	for _, resource := range apiDescription.Resources {
		for _, handler := range []*APIHandler{resource.Anon, resource.Auth} {
			if handler != nil {
				for i, action := range handler.Actions {
					doc, docParams, docReturns, err := action.parseRawDoc()
					if err != nil {
						return nil, "", fmt.Errorf("cannot parse action %q doc: %v", action.Name, err)
					}
					handler.Actions[i].Doc = doc
					handler.Actions[i].DocParams = docParams
					handler.Actions[i].DocReturns = docReturns
				}
			}
		}
	}

	rawIndentedJSON, err := json.MarshalIndent(rawJSON, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("cannot marshal raw indented JSON: %v", err)
	}
	return &apiDescription, string(rawIndentedJSON), nil
}

func (a *APIDescription) Format() string {
	var output string

	output += fmt.Sprintf("Doc: %q\n", a.Doc)
	output += fmt.Sprintf("Hash: %q\n", a.Hash)
	output += fmt.Sprintf("Resources:\n")
	resourcesByPath := a.Resources
	sort.Sort(apiResourcesByPath(resourcesByPath))
	for _, resource := range resourcesByPath {
		output += resource.Format()
	}

	return output
}

func (a APIResource) Format() string {
	var output string

	output += fmt.Sprintf("  %q:\n", a.Name)
	output += a.Anon.Format("    ", " (Anon)")
	output += a.Auth.Format("    ", " (Auth)")

	return output
}

func (a *APIHandler) Format(prefix, suffix string) string {
	if a == nil {
		return ""
	}

	var output string
	output += fmt.Sprintf("%s%q%s:\n", prefix, a.Name, suffix)
	output += fmt.Sprintf("%s  Doc: %q\n", prefix, a.Doc)
	output += fmt.Sprintf("%s  URI: %q\n", prefix, a.URI)
	output += fmt.Sprintf("%s  Actions:\n", prefix)

	for _, action := range a.Actions {
		output += fmt.Sprintf("%s    Name: %q\n", prefix, action.Name)
		output += fmt.Sprintf("%s    Doc: %q\n", prefix, action.Doc)
		if len(action.DocParams) == 0 {
			output += fmt.Sprintf("%s    DocParams: N/A\n", prefix)
		} else {
			output += fmt.Sprintf("%s    DocParams:\n", prefix)
			for paramName, param := range action.DocParams {
				output += fmt.Sprintf("%s      %q:\n", prefix, paramName)
				output += fmt.Sprintf("%s        Doc: %q\n", prefix, param.Doc)
				output += fmt.Sprintf("%s        PythonType: %v\n", prefix, param.PythonType)
				output += fmt.Sprintf("%s        GoType: %v\n", prefix, param.GoType)
			}
		}

		if len(action.DocReturns) == 0 {
			output += fmt.Sprintf("%s    DocReturns: N/A\n", prefix)
		} else {
			output += fmt.Sprintf("%s    DocReturns:\n", prefix)
			for code, codeDocs := range action.DocReturns {
				if len(codeDocs) == 1 {
					output += fmt.Sprintf("%s      %d: %s\n", prefix, code, codeDocs[0])
				} else {
					output += fmt.Sprintf("%s      %d:\n", prefix, code)
					for _, codeDoc := range codeDocs {
						output += fmt.Sprintf("%s      %s\n", prefix, codeDoc)
					}
				}
			}
		}

		if action.Op != nil {
			output += fmt.Sprintf("%s    Op: %q\n", prefix, *action.Op)
		} else {
			output += fmt.Sprintf("%s    Op: N/A\n", prefix)
		}

		output += fmt.Sprintf("%s    Restful: %v\n", prefix, action.Restful)
		output += fmt.Sprintf("%s    Method: %q\n", prefix, action.Method)
		output += fmt.Sprintf("%s  \n", prefix)
	}

	if len(a.Params) > 0 {
		paramsList := `"` + strings.Join(a.Params, `", "`) + `"`
		output += fmt.Sprintf("%s  Params: %s\n", prefix, paramsList)
	} else {
		output += fmt.Sprintf("%s  Params: N/A\n", prefix)
	}
	output += fmt.Sprintf("%s  Path: %q\n\n", prefix, a.Path)

	return output
}
