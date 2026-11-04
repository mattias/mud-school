// mudschool är kursverktyget: det håller reda på vilken lektion du är på,
// visar instruktioner och ledtrådar, och kör lektionens tester mot din kod.
//
// Verktyget använder bara Gos standardbibliotek, så det fungerar direkt
// efter `git clone`.
package main

import (
	"fmt"
	"os"
	"strings"
)

const usage = `mudschool – lär dig Go och Rust genom att bygga en MUD

Användning:
  mudschool init                 skapa din arbetsyta (workspace/) från startkoden
  mudschool list                 visa alla lektioner och din status
  mudschool show [id]            visa instruktionerna för aktuell (eller angiven) lektion
  mudschool hint                 visa nästa ledtråd för aktuell lektion
  mudschool check [--race]       kör testerna för aktuell lektion
  mudschool watch [--race]       kör testerna automatiskt varje gång du sparar
  mudschool goto <id>            hoppa till en lektion
  mudschool solution [--apply]   visa (eller kopiera in) referenslösningen för aktuell lektion
  mudschool run server|client    starta din server eller klient
  mudschool verify [id...]       (för lärare) kontrollera att alla lösningar klarar testerna
  mudschool reference <kat> [id] (för lärare) bygg referenslösningen till och med en lektion
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		return
	}
	root, err := findRoot()
	if err != nil {
		fail(err)
	}
	c := &course{root: root}
	if err := c.load(); err != nil {
		fail(err)
	}

	args := os.Args[2:]
	switch os.Args[1] {
	case "init":
		err = c.cmdInit()
	case "list", "ls":
		err = c.cmdList()
	case "show":
		err = c.cmdShow(args)
	case "hint":
		err = c.cmdHint()
	case "check":
		_, err = c.cmdCheck(hasFlag(args, "--race"), true)
	case "watch":
		err = c.cmdWatch(hasFlag(args, "--race"))
	case "goto":
		err = c.cmdGoto(args)
	case "solution":
		err = c.cmdSolution(hasFlag(args, "--apply"))
	case "run":
		err = c.cmdRun(args)
	case "verify":
		err = c.cmdVerify(args)
	case "reference":
		err = c.cmdReference(args)
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		err = fmt.Errorf("okänt kommando %q\n\n%s", os.Args[1], usage)
	}
	if err != nil {
		fail(err)
	}
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func fail(err error) {
	msg := err.Error()
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	fmt.Fprint(os.Stderr, red("fel: ")+msg)
	os.Exit(1)
}
