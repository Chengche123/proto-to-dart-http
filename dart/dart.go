package dart

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/xerrors"
)

func DartFileName(name string) string {
	return fmt.Sprintf("%s.dart", name)
}

type APIParam struct {
	ServiceName string
	HTTPMethod  string
	APIName     string
	Path        string
	Body        string
	FileName    string
	Request     Request
	Response    Response
}

func FileNames(as []*APIParam) []string {
	var names []string
	m := make(map[string]struct{})

	for i := range as {
		a := as[i]
		_, ok := m[a.FileName]
		if !ok {
			m[a.FileName] = struct{}{}
			names = append(names, a.FileName)
		}

		_, ok = m[a.Response.FileName]
		if !ok {
			m[a.Response.FileName] = struct{}{}
			names = append(names, a.Response.FileName)
		}

		_, ok = m[a.Request.FileName]
		if !ok {
			m[a.Request.FileName] = struct{}{}
			names = append(names, a.Request.FileName)
		}
	}
	return names
}

type Request struct {
	Name     string
	FileName string
}

type Response struct {
	Name     string
	FileName string
}

type GenerateDart struct {
	File *os.File
}

func NewGenerateDart(name string) (*GenerateDart, error) {
	fName := ProjectFileName(name)
	ext := filepath.Ext(fName)
	prefix := fName[:len(fName)-len(ext)]

	file, err := os.OpenFile(fmt.Sprintf("%s.pb.http.dart", prefix), os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		return nil, xerrors.Errorf("failed to get absolute path: %w", err)
	}

	return &GenerateDart{File: file}, nil
}

func WriteImports(g *GenerateDart, apiParams []*APIParam, project, path string) error {
	_, err := fmt.Fprint(g.File, "import 'dart:convert';\n")
	if err != nil {
		return xerrors.Errorf(": %w", err)
	}
	_, err = fmt.Fprint(g.File, "import 'package:http/http.dart' as http;\n")
	if err != nil {
		return xerrors.Errorf(": %w", err)
	}

	dartProject := ProjectFileName(project)
	files := FileNames(apiParams)
	for i := range files {
		file := files[i]
		sliceFile := strings.Split(file, "/")
		dartFile := strings.ReplaceAll(sliceFile[len(sliceFile)-1], "proto", "pb")
		if _, err := fmt.Fprintf(g.File, "import 'package:%s%s%s.dart';\n", dartProject, path, dartFile); err != nil {
			return xerrors.Errorf(": %w", err)
		}
	}

	return nil
}

//https://github.com/iancoleman/strcase

var numberSequence = regexp.MustCompile(`([a-zA-Z])(\d+)([a-zA-Z]?)`)
var numberReplacement = []byte(`$1 $2 $3`)

func addWordBoundariesToNumbers(s string) string {
	b := []byte(s)
	b = numberSequence.ReplaceAll(b, numberReplacement)
	return string(b)
}

func toCamelInitCase(s string, initCase bool) string {
	s = addWordBoundariesToNumbers(s)
	s = strings.Trim(s, " ")
	n := ""
	capNext := initCase
	for _, v := range s {
		if v >= 'A' && v <= 'Z' {
			n += string(v)
		}
		if v >= '0' && v <= '9' {
			n += string(v)
		}
		if v >= 'a' && v <= 'z' {
			if capNext {
				n += strings.ToUpper(string(v))
			} else {
				n += string(v)
			}
		}
		if v == '_' || v == ' ' || v == '-' {
			capNext = true
		} else {
			capNext = false
		}
	}
	return n
}

func toCamel(s string) string {
	return toCamelInitCase(s, true)
}

func WriteClass(g *GenerateDart, apiParams []*APIParam, project string) error {
	serviceName := apiParams[0].ServiceName
	_, err := fmt.Fprintf(g.File, "class %sClient {\n", serviceName)
	if err != nil {
		return xerrors.Errorf(": %w", err)
	}

	_, err = fmt.Fprint(g.File, "\tString baseUrl;\n")
	if err != nil {
		return xerrors.Errorf(": %w", err)
	}

	_, err = fmt.Fprintf(g.File, "\t%sClient(String baseUrl) {this.baseUrl = baseUrl;}\n", serviceName)
	if err != nil {
		return xerrors.Errorf(": %w", err)
	}

	for i := range apiParams {
		apiParam := apiParams[i]
		_, err = fmt.Fprintf(g.File,
			"\tFuture<%s> %c%s(%s body, Map<String, String> headers) async {\n"+
				"\t\tfinal response = await http.%s(\n"+
				"\t\t\tUri.parse(this.baseUrl + \"%s\"),\n"+
				"\t\t\tbody: json.encode(body),\n"+
				"\t\t\theaders: headers);\n\n"+
				"\t\tif (response.statusCode != 200) throw response.body;\n"+
				"\t\tvar raw = json.decode(Utf8Decoder().convert(response.bodyBytes));\n"+
				"\t\tfinal %s res = %s.fromJson(raw);\n"+
				"\t\treturn res;\n\t}\n\n",
			apiParam.Response.Name,
			strings.ToLower(apiParam.APIName)[0],
			apiParam.APIName[1:],
			apiParam.Request.Name,
			strings.ToLower(apiParam.HTTPMethod),
			apiParam.Path,
			apiParam.Response.Name,
			apiParam.Response.Name,
		)
		if err != nil {
			return xerrors.Errorf(": %w", err)
		}
	}

	_, err = fmt.Fprint(g.File, "}\n")
	if err != nil {
		return xerrors.Errorf(": %w", err)
	}

	return nil
}

func Build(apiParams []*APIParam, project, path string) (*GenerateDart, error) {
	if len(apiParams) < 1 {
		return nil, xerrors.Errorf("invalid apiParams")
	}

	g, err := NewGenerateDart(apiParams[0].FileName)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}

	if err := WriteImports(g, apiParams, project, path); err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}

	if err := WriteClass(g, apiParams, project); err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}

	return g, nil
}

func ProjectFileName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}
