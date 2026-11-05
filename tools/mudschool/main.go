// mudschool is the course tool: it keeps track of which lesson you are on,
// shows instructions and hints, and runs the lesson's tests against your code.
//
// It only uses the Go standard library, so it works right after `git clone`.
package main

import (
	"fmt"
	"os"
	"strings"
)

func usage() string {
	return L(
		`mudschool – lär dig Go och Rust genom att bygga en MUD

Användning:
  mudschool doctor               kontrollera att Go, Rust m.m. är installerade
  mudschool init [--lang en|sv|ja] skapa din arbetsyta (workspace/) från startkoden
  mudschool lang <en|sv|ja>      byt språk för kursen och spelets texter
  mudschool update               hämta kursens rättelser till din arbetsyta (din kod rörs inte)
  mudschool list                 visa alla lektioner och din status
  mudschool show [id]            visa instruktionerna för aktuell (eller angiven) lektion
  mudschool hint                 visa nästa ledtråd för aktuell lektion
  mudschool check [--race]       kör testerna för aktuell lektion
  mudschool watch [--race]       kör testerna automatiskt varje gång du sparar
  mudschool goto <id>            hoppa till en lektion
  mudschool solution [--apply]   visa (eller kopiera in) referenslösningen
  mudschool run server|client    starta din server eller klient
  mudschool verify [id...]       (lärare) kontrollera att alla lösningar klarar testerna
  mudschool reference <kat> [id] (lärare) bygg referenslösningen till och med en lektion
`,
		`mudschool – learn Go and Rust by building a MUD

Usage:
  mudschool doctor               check that Go, Rust etc. are installed
  mudschool init [--lang en|sv|ja] create your workspace (workspace/) from the starter code
  mudschool lang <en|sv|ja>      switch the language of the course and the game texts
  mudschool update               bring course fixes into your workspace (your code is not touched)
  mudschool list                 list all lessons and your progress
  mudschool show [id]            show the instructions for the current (or given) lesson
  mudschool hint                 show the next hint for the current lesson
  mudschool check [--race]       run the tests for the current lesson
  mudschool watch [--race]       run the tests automatically every time you save
  mudschool goto <id>            jump to a lesson
  mudschool solution [--apply]   show (or copy in) the reference solution
  mudschool run server|client    start your server or client
  mudschool verify [id...]       (teachers) check that every solution passes its tests
  mudschool reference <dir> [id] (teachers) build the reference solution up to a lesson
`,
		`mudschool – MUD を作りながら Go と Rust を学ぶ

使い方:
  mudschool doctor               Go や Rust などがインストールされているか確認する
  mudschool init [--lang en|sv|ja] スターターコードから作業領域 (workspace/) を作る
  mudschool lang <en|sv|ja>      コースとゲーム内テキストの言語を切り替える
  mudschool update               コースの修正を作業領域に取り込む（あなたのコードには触れない）
  mudschool list                 レッスン一覧と進捗を表示する
  mudschool show [id]            現在の（または指定した）レッスンの説明を表示する
  mudschool hint                 現在のレッスンの次のヒントを表示する
  mudschool check [--race]       現在のレッスンのテストを実行する
  mudschool watch [--race]       保存するたびに自動でテストを実行する
  mudschool goto <id>            指定したレッスンへ移動する
  mudschool solution [--apply]   模範解答を表示する（またはコピーする）
  mudschool run server|client    自分のサーバーまたはクライアントを起動する
  mudschool verify [id...]       （講師向け）すべての模範解答がテストに通るか確認する
  mudschool reference <dir> [id] （講師向け）指定レッスンまでの模範解答を作る
`)
}

func main() {
	args, langFlag := extractLangFlag(os.Args[1:])
	lang = detectLang()

	root, err := findRoot()
	if err != nil {
		fail(err)
	}
	c := &course{root: root}
	// The language saved in the workspace wins over the environment.
	if p, err := c.loadProgress(); err == nil && validLang(p.Lang) {
		lang = p.Lang
	}
	if langFlag != "" {
		if !validLang(langFlag) {
			fail(fmt.Errorf("--lang: %q (en, sv, ja)", langFlag))
		}
		lang = langFlag
	}
	if len(args) == 0 {
		fmt.Print(usage())
		return
	}
	if err := c.load(); err != nil {
		fail(err)
	}

	cmd, rest := args[0], args[1:]
	switch cmd {
	case "doctor":
		err = c.cmdDoctor()
	case "init":
		err = c.cmdInit()
	case "lang":
		err = c.cmdLang(rest)
	case "update":
		err = c.cmdUpdate()
	case "list", "ls":
		err = c.cmdList()
	case "show":
		err = c.cmdShow(rest)
	case "hint":
		err = c.cmdHint()
	case "check":
		_, err = c.cmdCheck(hasFlag(rest, "--race"), true)
	case "watch":
		err = c.cmdWatch(hasFlag(rest, "--race"))
	case "goto":
		err = c.cmdGoto(rest)
	case "solution":
		err = c.cmdSolution(hasFlag(rest, "--apply"))
	case "run":
		err = c.cmdRun(rest)
	case "verify":
		err = c.cmdVerify(rest, langFlag)
	case "reference":
		err = c.cmdReference(rest)
	case "help", "-h", "--help":
		fmt.Print(usage())
	default:
		err = fmt.Errorf("%s %q\n\n%s", L("okänt kommando", "unknown command", "不明なコマンド"), cmd, usage())
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
	fmt.Fprint(os.Stderr, red(L("fel: ", "error: ", "エラー: "))+msg)
	os.Exit(1)
}
