// Code generated for the course language: 日本語.
//
// All text the game shows to players lives in this file, so that the course
// can be taken in Swedish, English or Japanese. The tests use these constants
// too, so keep the names – but feel free to change the texts.
// `./mudschool lang <sv|en|ja>` replaces this file.
package mud

const (
	// the first thing a new player sees (lesson 1)
	Welcome = "MUD School へようこそ！"
	// login prompt (lesson 5)
	MsgAskName = "お名前は？"
	// after login: name
	MsgWelcomeName = "ようこそ、%s さん！"
	// bad name: error
	MsgInvalidName     = "その名前は使えません: %v。"
	MsgErrNameTooShort = "名前は 2 文字以上にしてください"
	MsgErrNameTooLong  = "名前は 16 文字以内にしてください"
	MsgErrNameInvalid  = "名前には文字だけを使ってください"
	MsgErrNameTaken    = "その名前はすでに使われています"
	// unknown verb, formatted with %q (lesson 4)
	MsgUnknownCommand = "不明なコマンド %q です。help と入力するとヘルプが表示されます。"
	// help: comma separated verbs
	MsgCommands = "コマンド: %s"
	// quit
	MsgGoodbye = "さようなら！"
	// who: count, names (lesson 5)
	MsgWho = "オンラインのプレイヤー (%d): %s"
	// to everyone else when a player logs in (lesson 7)
	MsgJoined = "%s さんが接続しました。"
	// to everyone else when a player leaves
	MsgLeft = "%s さんが去りました。"
	// say, to the speaker
	MsgYouSay = "あなた:「%s」"
	// say, to the others: name, text
	MsgSays = "%s:「%s」"
	// say without text
	MsgSayWhat = "何と言いますか？"
	// shout, to the shouter (lesson 8)
	MsgYouShout = "あなたは叫んだ:「%s」"
	// shout, to everyone else: name, text
	MsgShouts = "%s が叫んだ:「%s」"
	// shout without text
	MsgShoutWhat = "何と叫びますか？"
	// look: list of exits
	MsgExits = "出口: %s"
	// look: no exits
	MsgNoExits = "出口: なし"
	// look: other players in the room
	MsgHere = "ここにいる人: %s"
	// no exit in that direction
	MsgCantGo = "その方向には進めません。"
	// to the old room: name, directionTravel[dir]
	MsgLeaves = "%s は%sへ去った。"
	// to the new room: name
	MsgArrives = "%s がやって来た。"
	// separator in lists of names and exits
	MsgListSep = "、"
	// world file errors (lesson 9) – must keep %w
	MsgErrReadWorld     = "世界ファイルの読み込み: %w"
	MsgErrDuplicateRoom = "部屋 %q が 2 回定義されています"
	MsgErrStartMissing  = "開始部屋 %q が存在しません"
	MsgErrRoomFields    = "部屋 %q: id と name は必須です"
	MsgErrTerrain       = "部屋 %q: 不明な地形 %q"
	MsgErrSamePos       = "部屋 %q: 座標 %v が %q と重なっています"
	MsgErrExitMissing   = "部屋 %q: 出口 %s の行き先 %q が存在しません"
	MsgErrExitCoords    = "部屋 %q: 出口 %s は %q へ続きますが、座標が合いません"
	MsgErrDirection     = "部屋 %q: 不明な方向 %q"
)

// directionNames are the directions as shown to the player (lesson 8).
var directionNames = map[string]string{
	"north": "北",
	"east":  "東",
	"south": "南",
	"west":  "西",
	"up":    "上",
	"down":  "下",
}

// directionTravel is used in MsgLeaves (lesson 8).
var directionTravel = map[string]string{
	"north": "北",
	"east":  "東",
	"south": "南",
	"west":  "西",
	"up":    "上",
	"down":  "下",
}
