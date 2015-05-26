package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	flagMode string
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	flag.StringVar(&flagMode, "mode", "production", "deployment mode")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: convox/app [options]\n")
		fmt.Fprintf(os.Stderr, "  expects an optional docker-compose.yml on stdin\n\n")
		fmt.Fprintf(os.Stderr, "options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nexamples:\n")
		fmt.Fprintf(os.Stderr, "  cat docker-compose.yml | docker run -i convox/app -mode staging\n")
	}
}

type ManifestEntry struct {
	Command string   `yaml:"command"`
	Links   []string `yaml:"links"`
	Ports   []string `yaml:"ports"`
}

type Manifest map[string]ManifestEntry

type Listener struct {
	Balancer string
	Process  string
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", err)
	os.Exit(1)
}

func usage() {
	fmt.Fprintf(os.Stderr, `app: build convox app stacks

Options:
  -p <balancer:container>    map a port on the balancer to a container port
`)

	os.Exit(1)
}

func main() {
	flag.Parse()

	var manifest Manifest

	if stat, _ := os.Stdin.Stat(); stat.Mode()&os.ModeCharDevice == 0 {
		man, err := ioutil.ReadAll(os.Stdin)

		if err != nil {
			die(err)
		}

		err = yaml.Unmarshal(man, &manifest)

		if err != nil {
			die(err)
		}
	}

	data, err := buildTemplate(flagMode, "formation", manifest)

	if err != nil {
		displaySyntaxError(data, err)
		die(err)
	}

	pretty, err := prettyJson(data)

	if err != nil {
		displaySyntaxError(data, err)
		die(err)
	}

	fmt.Println(pretty)
}

func buildTemplate(name, section string, data interface{}) (string, error) {
	tmpl, err := template.New(section).Funcs(templateHelpers()).ParseFiles(fmt.Sprintf("template/%s.tmpl", name))

	if err != nil {
		return "", err
	}

	var formation bytes.Buffer

	err = tmpl.Execute(&formation, data)

	if err != nil {
		return "", err
	}

	return formation.String(), nil
}

func displaySyntaxError(data string, err error) {
	syntax, ok := err.(*json.SyntaxError)

	if !ok {
		fmt.Println(err)
		return
	}

	start, end := strings.LastIndex(data[:syntax.Offset], "\n")+1, len(data)

	if idx := strings.Index(data[start:], "\n"); idx >= 0 {
		end = start + idx
	}

	line, pos := strings.Count(data[:start], "\n"), int(syntax.Offset)-start-1

	fmt.Printf("Error in line %d: %s \n", line, err)
	fmt.Printf("%s\n%s^\n", data[start:end], strings.Repeat(" ", pos))
}

func parseList(list string) []string {
	return strings.Split(list, ",")
}

func parseListeners(list string) []Listener {
	items := parseList(list)

	listeners := make([]Listener, len(items))

	for i, l := range items {
		parts := strings.SplitN(l, ":", 2)

		if len(parts) != 2 {
			die(fmt.Errorf("listeners must be balancer:process pairs"))
		}

		listeners[i] = Listener{Balancer: parts[0], Process: parts[1]}
	}

	return listeners
}

func prettyJson(raw string) (string, error) {
	var parsed map[string]interface{}

	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return "", err
	}

	bp, err := json.MarshalIndent(parsed, "", "  ")

	if err != nil {
		return "", err
	}

	clean := strings.Replace(string(bp), "\n\n", "\n", -1)

	return clean, nil
}

func printLines(data string) {
	lines := strings.Split(data, "\n")

	for i, line := range lines {
		fmt.Printf("%d: %s\n", i, line)
	}
}

func templateHelpers() template.FuncMap {
	return template.FuncMap{
		"ingress": func(ps string, entry ManifestEntry) template.HTML {
			ls := []string{}

			for _, port := range entry.Ports {
				parts := strings.SplitN(port, ":", 2)

				if len(parts) != 2 {
					continue
				}

				ls = append(ls, fmt.Sprintf(`{ "CidrIp": "0.0.0.0/0", "IpProtocol": "tcp", "FromPort": { "Ref": "%sPort%sBalancer" }, "ToPort": { "Ref": "%sPort%sBalancer" } }`, upperName(ps), parts[0], upperName(ps), parts[0]))
			}

			return template.HTML(strings.Join(ls, ","))
		},
		"listeners": func(ps string, entry ManifestEntry) template.HTML {
			ls := []string{}

			for _, port := range entry.Ports {
				parts := strings.SplitN(port, ":", 2)

				if len(parts) != 2 {
					continue
				}

				ls = append(ls, fmt.Sprintf(`{ "Protocol": "TCP", "LoadBalancerPort": { "Ref": "%sPort%sBalancer" }, "InstanceProtocol": "TCP", "InstancePort": { "Ref": "%sPort%sHost" } }`, upperName(ps), parts[0], upperName(ps), parts[0]))
			}

			return template.HTML(strings.Join(ls, ","))
		},
		"links": func(e ManifestEntry) template.HTML {
			ls := []string{}

			for _, link := range e.Links {
				ls = append(ls, fmt.Sprintf(`"%s"`, link))
			}

			return template.HTML(strings.Join(ls, ","))
		},
		"loadbalancers": func(ps string, e ManifestEntry) template.HTML {
			ls := []string{}

			for _, port := range e.Ports {
				parts := strings.SplitN(port, ":", 2)

				if len(parts) != 2 {
					continue
				}

				ls = append(ls, fmt.Sprintf(`{ "Fn::Join": [ ":", [ { "Ref": "%sBalancer" }, "%s" ] ] }`, upperName(ps), parts[1]))
			}

			return template.HTML(strings.Join(ls, ","))
		},
		"portmappings": func(ps string, e ManifestEntry) template.HTML {
			ls := []string{}

			for _, port := range e.Ports {
				parts := strings.SplitN(port, ":", 2)

				if len(parts) != 2 {
					continue
				}

				ls = append(ls, fmt.Sprintf(`{ "Fn::Join": [ ":", [ { "Ref": "%sPort%sHost" }, "%s" ] ] }`, upperName(ps), parts[0], parts[1]))
			}

			// for _, l := range listeners {
			//   if l.Process == process {
			//     b := upperName(l.Balancer)
			//     p := upperName(l.Process)
			//     ls = append(ls, fmt.Sprintf(`{ "Fn::Join": [ ":", [ { "Ref": "%s%sHostPort" }, { "Ref": "%s%sContainerPort" } ] ] }`, b, p, b, p))
			//   }
			// }

			return template.HTML(strings.Join(ls, ","))
		},
		"randomport": func() int {
			return rand.Intn(50000) + 5000
		},
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
		"split": func(ss string, t string) []string {
			return strings.Split(ss, t)
		},
		"upper": func(name string) string {
			return upperName(name)
		},
	}
}

func upperName(name string) string {
	us := strings.ToUpper(name[0:1]) + name[1:]

	for {
		i := strings.Index(us, "-")

		if i == -1 {
			break
		}

		s := us[0:i]

		if len(us) > i+1 {
			s += strings.ToUpper(us[i+1 : i+2])
		}

		if len(us) > i+2 {
			s += us[i+2:]
		}

		us = s
	}

	return us
}

func (me ManifestEntry) FirstPort() string {
	if len(me.Ports) > 0 {
		return strings.Split(me.Ports[0], ":")[0]
	}

	return ""
}

func (me ManifestEntry) HasPorts() bool {
	return len(me.Ports) > 0
}
