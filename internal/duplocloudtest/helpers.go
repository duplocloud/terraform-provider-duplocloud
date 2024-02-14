package duplocloudtest

import (
	"strings"
)

func WriteCustomResource(rType, rName string, writer func(*strings.Builder)) string {
	var sb strings.Builder
	sb.WriteString("resource \"")
	sb.WriteString(rType)
	sb.WriteString("\" \"")
	sb.WriteString(rName)
	sb.WriteString("\" {\n")
	writer(&sb)
	sb.WriteString("}\n")
	return sb.String()
}

func WriteFlatResource(rType, rName string, defaults, overrides map[string]string) string {
	var sb strings.Builder
	sb.WriteString("resource \"")
	sb.WriteString(rType)
	sb.WriteString("\" \"")
	sb.WriteString(rName)
	sb.WriteString("\" {\n")
	writeAttrs(&sb, defaults, overrides)
	sb.WriteString("}\n")
	return sb.String()
}

func WriteAttrs(defaults, overrides map[string]string) string {
	var sb strings.Builder
	writeAttrs(&sb, defaults, overrides)
	return sb.String()
}

func writeAttrs(sb *strings.Builder, defaults, overrides map[string]string) {
	for k, v := range defaults {
		if _, ok := overrides[k]; !ok {
			WriteAttr(sb, k, v)
		}
	}
	for k, v := range overrides {
		WriteAttr(sb, k, v)
	}
}

func WriteAttr(sb *strings.Builder, key, value string) {
	sb.WriteString("  ")
	sb.WriteString(key)
	sb.WriteString(" = ")
	sb.WriteString(value)
	sb.WriteString("\n")
}
