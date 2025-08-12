package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const (
	_MAX_TEMPLATE_LENGTH      = 1000000 // 1MB
	_MAX_VARIABLE_LENGTH      = 1000000 // 1MB
	_MAX_TEMPLATE_EXPRESSIONS = 10000
	_MAX_REGEX_LENGTH         = 1000000 // 1MB
)

type uriTemplatePartType string

var valueUriTemplatePartType uriTemplatePartType
var TextUriTemplatePartType uriTemplatePartType
var uriTemplateOperators = []string{"+", "#", ".", "/", "?", "&"}

type uriTemplatePart struct {
	name     string
	operator string
	names    []string
	exploded bool
	partType uriTemplatePartType
}

type UriTemplate struct {
	template string
	parts    []uriTemplatePart
}

type UriVariables = map[string][]string

func NewUriTemplate(template string) (*UriTemplate, error) {
	ut := new(UriTemplate)
	err := ut.validateLength(template, _MAX_TEMPLATE_LENGTH, "Template")
	if err != nil {
		return nil, fmt.Errorf("ut.validateLength %v", err)
	}
	ut.template = template
	ut.parts, err = ut.parse(template)
	if err != nil {
		return nil, fmt.Errorf("ut.parse %v", err)
	}
	return ut, nil
}

func (ut *UriTemplate) validateLength(str string, max int, ctx string) error {
	if len(str) > max {
		return fmt.Errorf("%s exceeds maximum length of %d characters (got %d)", ctx, max, len(str))
	}
	return nil
}

func (ut *UriTemplate) getOperator(expr string) string {
	for _, op := range uriTemplateOperators {
		if !strings.HasPrefix(expr, string(op)) {
			continue
		}
		return op
	}
	return ""
}

func (ut *UriTemplate) getNames(expr string, operator string) []string {
	op := operator
	if op == "" {
		op = ut.getOperator(expr)
	}
	rawLabels := strings.Split(expr[len(op):], ",")
	var labels []string
	for _, l := range rawLabels {
		//Remove "*" and trim spaces
		cleanL := strings.TrimSpace(strings.ReplaceAll(l, "*", ""))
		if len(cleanL) == 0 {
			continue
		}
		labels = append(labels, cleanL)
	}
	return labels
}

func (ut *UriTemplate) encodeValue(value string, operator string) (string, error) {
	err := ut.validateLength(value, _MAX_VARIABLE_LENGTH, "Variable value")
	if err != nil {
		return "", fmt.Errorf("ut.validateLength %v", err)
	}
	if operator == "+" || operator == "#" {
		//preserve reserved characters
		return url.PathEscape(value), nil
	}

	//encodeURIComponent: stricter encoding
	return url.QueryEscape(value), nil
}

func (ut *UriTemplate) parse(template string) ([]uriTemplatePart, error) {
	var parts []uriTemplatePart
	currentText := strings.Builder{}
	i := 0
	exprCount := 0
	for i < len(template) {
		if template[i] == '{' {
			if currentText.Len() > 0 {
				parts = append(parts, uriTemplatePart{name: currentText.String(), partType: TextUriTemplatePartType})
				currentText.Reset()
			}

			end := strings.IndexByte(template[i:], '}')
			if end == -1 {
				return nil, fmt.Errorf("unclosed template expression")
			}
			end += i

			exprCount++
			if exprCount > _MAX_TEMPLATE_EXPRESSIONS {
				return nil, fmt.Errorf("template contains too many expressions")
			}

			expr := template[i+1 : end]
			operator := ut.getOperator(expr)
			exploded := strings.Contains(expr, "*")
			names := ut.getNames(expr, operator)
			name := names[0]

			for _, n := range names {
				if err := ut.validateLength(n, _MAX_VARIABLE_LENGTH, "Variable name"); err != nil {
					return nil, err
				}
			}

			parts = append(parts, uriTemplatePart{
				name:     name,
				operator: operator,
				names:    names,
				exploded: exploded,
				partType: valueUriTemplatePartType,
			})

			i = end + 1
		} else {
			currentText.WriteByte(template[i])
			i++
		}
	}

	if currentText.Len() > 0 {
		parts = append(parts, uriTemplatePart{name: currentText.String(), partType: TextUriTemplatePartType})
	}

	return parts, nil
}

func (ut *UriTemplate) expandPart(part uriTemplatePart, variables UriVariables) (string, error) {
	var pairs []string
	if part.operator == "?" || part.operator == "&" {
		for _, name := range part.names {
			variable, okVariable := variables[name]
			if !okVariable {
				continue
			}
			var encodes []string
			for _, ev := range variable {
				encode, err := ut.encodeValue(ev, part.operator)
				if err == nil {
					return "", fmt.Errorf("ut.encodeValue variable value:%s, %v", ev, err)
				}
				encodes = append(encodes, encode)
			}
			expand := fmt.Sprintf("%s=%s", name, strings.Join(encodes, ","))
			pairs = append(pairs, expand)
		}

		if len(pairs) == 0 {
			return "", nil
		}

		separator := part.operator
		result := separator + strings.Join(pairs, "&")
		return result, nil
	}
	if len(part.names) > 1 {
		for _, name := range part.names {
			variable, okVariable := variables[name]
			if !okVariable {
				continue
			}
			pairs = append(pairs, variable[0])
		}
		if len(pairs) == 0 {
			return "", nil
		}
		return strings.Join(pairs, ","), nil
	}

	variable, okVariable := variables[part.name]
	if !okVariable {
		return "", nil
	}
	var encodes []string
	for _, ev := range variable {
		encode, err := ut.encodeValue(ev, part.operator)
		if err == nil {
			return "", fmt.Errorf("ut.encodeValue variable value:%s, %v", ev, err)
		}
		encodes = append(encodes, encode)
	}
	switch part.operator {
	case "":
	case "+":
		return strings.Join(encodes, ","), nil
	case "#":
		return "#" + strings.Join(encodes, ","), nil
	case ".":
		return "." + strings.Join(encodes, "."), nil
	case "/":
		return "/" + strings.Join(encodes, "/"), nil
	}
	return strings.Join(encodes, ","), nil
}

func (ut *UriTemplate) escapeRegExp(str string) string {
	return regexp.QuoteMeta(str)
}

func (ut *UriTemplate) partToRegExp(part uriTemplatePart) (map[string]string, error) {
	patterns := make(map[string]string)
	for _, name := range part.names {
		err := ut.validateLength(name, _MAX_VARIABLE_LENGTH, "Variable name")
		if err != nil {
			return nil, fmt.Errorf("ut.validateLength part.name:%s,%v", name, err)
		}
	}

	if part.operator == "?" || part.operator == "&" {
		prefix := "\\" + part.operator
		for _, name := range part.names {
			escapeRegExp := ut.escapeRegExp(name)
			patterns[name] = prefix + escapeRegExp + "=([^&]+)"
			prefix = "&"
		}
		return patterns, nil
	}

	var pattern string
	switch part.operator {
	case "":
		pattern = "([^/,]+)"
		if part.exploded {
			pattern = "([^/]+(?:,[^/]+)*)"
		}
	case "+":
	case "#":
		pattern = "(.+)"
	case ".":
		pattern = "\\.([^/,]+)"
	case "/":
		pattern = "/"
		if part.exploded {
			pattern += "([^/]+(?:,[^/]+)*)"
		} else {
			pattern += "([^/,]+)"
		}
	default:
		pattern = "([^/]+)"
	}
	patterns[part.name] = pattern
	return patterns, nil
}

//Returns true if the given string contains any URI template expressions.
//A template expression is a sequence of characters enclosed in curly braces,
//like {foo} or {?bar}.
func (ut *UriTemplate) IsTemplate(str string) bool {
	//Look for any sequence of characters between curly braces
	//that isn't just whitespace
	re := regexp.MustCompile(`\{[^}\s]+\}`)
	return re.MatchString(str)
}

func (ut *UriTemplate) GetVariableNames() []string {
	var r []string
	for _, p := range ut.parts {
		if p.partType == TextUriTemplatePartType {
			continue
		}
		r = append(r, p.names...)
	}
	return r
}

func (ut *UriTemplate) String() string {
	return ut.template
}

func (ut *UriTemplate) Expand(variables UriVariables) (string, error) {
	var result string
	var hasQueryParam bool

	for _, part := range ut.parts {
		if part.partType == TextUriTemplatePartType {
			result += part.name
			continue
		}
		expanded, err := ut.expandPart(part, variables)
		if err != nil {
			return "", fmt.Errorf("ut.expandPart part.name:%s, %v", part.name, err)
		}
		if expanded == "" {
			continue
		}
		//Convert ? to & if we already have a query parameter
		if (part.operator == "?" || part.operator == "&") && hasQueryParam {
			result += strings.ReplaceAll(expanded, "?", "&")
		} else {
			result += expanded
		}

		hasQueryParam = part.operator == "?" || part.operator == "&"
	}
	return result, nil
}

func (ut *UriTemplate) Match(uri string) (UriVariables, error) {
	err := ut.validateLength(uri, _MAX_TEMPLATE_LENGTH, "URI")
	if err != nil {
		return nil, fmt.Errorf("ut.validateLength URI, %v", err)
	}
	pattern := "^"
	namesExploded := make(map[string]bool)
	for _, part := range ut.parts {
		if part.partType == TextUriTemplatePartType {
			pattern += ut.escapeRegExp(part.name)
			continue
		}
		patterns, err := ut.partToRegExp(part)
		if err != nil {
			return nil, fmt.Errorf("ut.partToRegExp part.name: %s, %v", part.name, err)
		}
		for name, partPattern := range patterns {
			pattern += partPattern
			namesExploded[name] = part.exploded
		}
	}
	pattern += "$"
	err = ut.validateLength(pattern, _MAX_REGEX_LENGTH, "Generated regex pattern")
	if err != nil {
		return nil, fmt.Errorf("ut.validateLength Generated regex pattern, %v", err)
	}
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("regexp.Compile, %v", err)
	}
	match := regex.FindAllString(uri, -1)
	if match == nil {
		return nil, nil
	}
	result := UriVariables{}
	var nameIndex int
	for name, exploded := range namesExploded {
		value := match[nameIndex+1]
		cleanName := strings.ReplaceAll(name, "*", "")
		if exploded && strings.Contains(value, ",") {
			result[cleanName] = strings.Split(value, ",")
		} else {
			result[cleanName] = []string{value}
		}
		nameIndex++
	}
	return result, nil
}
