// Code generated for the course language: 日本語.
//
// All text the game shows to players lives in this file, so that the course
// can be taken in Swedish, English or Japanese. The tests use these constants
// too, so keep the names – but feel free to change the texts.
// `./mudschool lang <en|sv|ja>` replaces this file.
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
	// time: day, hour, minute (lesson 14)
	MsgTime = "%d 日目、%02d:%02d です。"
	// to everyone when the server stops
	MsgShutdown = "サーバーを停止します。またお会いしましょう！"
	// look: items in the room (lesson 15)
	MsgItemsHere = "ここにある物: %s。"
	// get without an argument
	MsgGetWhat = "何を取りますか？"
	// drop without an argument
	MsgDropWhat = "何を置きますか？"
	// examine without an argument
	MsgExamineWhat = "何を調べますか？"
	// nothing matches the word here
	MsgNoSuchItem = "ここに %q は見当たりません。"
	// drop: not in the inventory
	MsgNotCarrying = "%q は持っていません。"
	// get: item name
	MsgYouGet = "%s を拾った。"
	// get, to the others: player, item
	MsgGets = "%s が %s を拾った。"
	// drop: item name
	MsgYouDrop = "%s を置いた。"
	// drop, to the others: player, item
	MsgDrops = "%s が %s を置いた。"
	// inventory: list of items
	MsgInventory = "持ち物: %s。"
	// inventory: nothing
	MsgInventoryEmpty = "何も持っていません。"
	// world file: items (lesson 15)
	MsgErrItemFields    = "アイテム %q: id と name は必須です"
	MsgErrDuplicateItem = "アイテム %q が 2 回定義されています"
	MsgErrItemRoom      = "アイテム %q: 部屋 %q が存在しません"
	// world file: creatures (lesson 16)
	MsgErrDuplicateNPC = "生き物 %q が 2 回定義されています"
	MsgErrNPCRoom      = "生き物 %q: 部屋 %q が存在しません"
	MsgErrNPCFields    = "生き物 %q: id と name は必須です"
	MsgErrBehavior     = "生き物 %q: 不明な振る舞い %q"
	// kill without an argument (lesson 18)
	MsgAttackWhat = "何を攻撃しますか？"
	// no creature matches
	MsgNoSuchTarget = "ここに戦える %q はいません。"
	// the creature is peaceful
	MsgCantAttack = "%s は攻撃できません。"
	// you start a fight: creature
	MsgYouAttack = "%s に攻撃を仕掛けた！"
	// to the others: player, creature
	MsgAttacks = "%s が %s に攻撃を仕掛けた！"
	// your hit: creature, damage
	MsgYouHit = "%s に命中（%d ダメージ）。"
	// the creature hits you: creature, damage
	MsgHitsYou = "%s の攻撃が命中（%d ダメージ）。"
	// you won: creature
	MsgYouKill = "%s を倒した！"
	// to the others: player, creature
	MsgKills = "%s が %s を倒した！"
	// you died: the room you wake up in
	MsgYouDie = "あなたは倒れた……そして %s で目を覚ました。"
	// to the others: player
	MsgDies = "%s は倒れ、姿を消した。"
	// a creature comes back: creature
	MsgAppears = "%s が現れた。"
	// flee when not in a fight
	MsgNotFighting = "誰とも戦っていません。"
	// flee with no exits
	MsgFleeNowhere = "逃げ場がない！"
	// you fled: direction
	MsgYouFlee = "%sへ逃げ出した！"
	// login with saving (lesson 20)
	MsgAskPassword = "パスワード:"
	// new player: name
	MsgChoosePassword   = "%s のパスワードを決めてください:"
	MsgPasswordTooShort = "パスワードは 4 文字以上にしてください。"
	MsgWrongPassword    = "パスワードが違います。"
	MsgStoreError       = "保存で問題が起きました。しばらくしてからもう一度試してください。"
	// zones (lesson 23)
	MsgBlocked     = "そこへは歩けません。"
	MsgNoStairs    = "ここに階段はありません。"
	MsgGoWhere     = "どこへ行きますか？（例: go north）"
	MsgErrZoneRows = "部屋 %q: ゾーンが正しくありません（%v）。3×3 以上、同じ長さの行、既知のタイルのみ"
	MsgErrZoneExit = "部屋 %q: 出口 %s に対応する開口部がゾーンにありません"
	// creatures that hunt (lesson 27)
	MsgAttacksYou = "%s が襲いかかってきた！"
	// reloading the world (lesson 31)
	MsgNotAdmin     = "それは管理者だけができます。"
	MsgNoWorldFile  = "サーバーはワールドファイル（-world）なしで起動しました。再読み込みするものがありません。"
	MsgReloadFailed = "ワールドファイルに誤りがあります。何も変更していません:\n%v"
	MsgReloaded     = "世界が揺らめいた。作り直されたのだ。"
	MsgRoomGone     = "いた部屋がなくなった。気づくと別の場所にいる。"
	// scripts in Lua (lesson 32)
	MsgScriptError      = "スクリプトで問題が起きました。"
	MsgScriptLoadFailed = "スクリプトに誤りがあります:\n%v"
	// plugins in WebAssembly (lesson 33)
	MsgPluginError      = "プラグインで問題が起きました。"
	MsgPluginLoadFailed = "プラグインに誤りがあります:\n%v"
	// events and hooks (lesson 34)
	MsgCancelled = "何かに阻まれた。"
	// copyover (lesson 35)
	MsgCopyoverDone   = "一瞬、世界が震えた……すべては元のまま、それでいて新しい。"
	MsgCopyoverFailed = "copyover に失敗しました: %v"
	// building in the game (lesson 36)
	MsgDigUsage     = "使い方: @dig <方向> <名前>（例: @dig north 暗い洞窟）"
	MsgDescUsage    = "使い方: @desc <説明>"
	MsgNameUsage    = "使い方: @name <名前>"
	MsgExitUsage    = "使い方: @exit <方向> <部屋 ID>（消すときは none）"
	MsgExitExists   = "%sにはすでに出口があります。"
	MsgNewRoomDesc  = "何もない部屋。@desc で説明を書こう。"
	MsgDug          = "%sへ掘り進み、%q（%s）を作った。"
	MsgBuilt        = "完了。"
	MsgBuildInvalid = "それでは世界が正しくなくなります。何も変更していません:\n%v"
	MsgNotSaved     = "（サーバーにワールドファイルがないので、変更は保存されません。）"
	MsgSaveFailed   = "世界を保存できませんでした: %v"
	// the tutorial (lesson 37)
	MsgTutorialStep     = "[チュートリアル %d/%d] %s"
	MsgTutorialSkipped  = "チュートリアルを抜けました。help ですべてのコマンドが見られます。"
	MsgTutorialNotIn    = "チュートリアル中ではありません。"
	MsgDisappears       = "%s が消えた。"
	MsgErrTutorialStart = "チュートリアル: 開始の部屋 %q がありません"
	MsgErrTutorialOpen  = "チュートリアル: チュートリアルから世界の開始の部屋 %q に行けてしまいます。チュートリアルの区域は閉じていなければなりません"
	MsgErrTutorialEmpty = "チュートリアル: ステップがありません"
	MsgErrTutorialStep  = "チュートリアルのステップ %d: テキストと、command・room・item のちょうど 1 つが必要です"
	MsgErrTutorialRoom  = "チュートリアルのステップ %d: 部屋 %q はチュートリアルの区域にありません"
	MsgErrTutorialItem  = "チュートリアルのステップ %d: 物 %q はチュートリアルの区域にありません"
	// languages per player (lesson 38)
	MsgLanguage        = "あなたの言語: %s。あるのは: %s。例えば language ja で選べます"
	MsgLanguageSet     = "言語: %s。"
	MsgLanguageUnknown = "%q のテキストはありません。あるのは: %s。"
	MsgErrText         = "テキストは文字列か、言語ごとに 1 つの文字列を持つオブジェクトでなければなりません"
	MsgErrTemplate     = "%q のテキストは正しいテンプレートではありません: %v"
	MsgErrTutorialText = "チュートリアルのステップ %d: %v"
	// protection (lesson 43)
	MsgSlowDown           = "落ち着いて – コマンドの送り方が速すぎます。"
	MsgFlooding           = "コマンドが多すぎます – さようなら。"
	MsgTooManyConnections = "あなたのアドレスからの接続が多すぎます – 後でもう一度どうぞ。"
	MsgLineTooLong        = "その行は長すぎます – さようなら。"
	MsgLoginTimeout       = "ログインに時間がかかりすぎました – さようなら。"
)

// CourseLanguage is the language this file – and world.json – is written in (lesson 38).
const CourseLanguage = "ja"

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
