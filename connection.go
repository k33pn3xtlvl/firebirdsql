/*******************************************************************************
The MIT License (MIT)

Copyright (c) 2013-2016 Hajime Nakagami

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*******************************************************************************/

package firebirdsql

import (
	"context"
	"database/sql/driver"
	"math/big"
)

type firebirdsqlConn struct {
	wp                *wireProtocol
	tx                *firebirdsqlTx
	addr              string
	dbName            string
	user              string
	password          string
	columnNameToLower bool
	isAutocommit      bool
	clientPublic      *big.Int
	clientSecret      *big.Int
	transHandles      []int32
	timeZoneIds       map[int]string
}

func (fc *firebirdsqlConn) begin(isolationLevel int) (driver.Tx, error) {
	tx, err := newFirebirdsqlTx(fc, isolationLevel, false)
	fc.tx = tx
	return driver.Tx(tx), err
}

func (fc *firebirdsqlConn) Begin() (driver.Tx, error) {
	return fc.begin(ISOLATION_LEVEL_READ_COMMITED)
}

func (fc *firebirdsqlConn) Close() (err error) {
	for _, h := range fc.transHandles {
		fc.wp.opRollback(h)
	}

	_, _, _, err = fc.wp.opResponse()
	if err != nil {
		return
	}
	fc.wp.opDetach()
	_, _, _, err = fc.wp.opResponse()
	fc.wp.conn.Close()
	return
}

func (fc *firebirdsqlConn) prepare(ctx context.Context, query string) (driver.Stmt, error) {
	return newFirebirdsqlStmt(fc, query)
}

func (fc *firebirdsqlConn) Prepare(query string) (driver.Stmt, error) {
	return fc.prepare(context.Background(), query)
}

func (fc *firebirdsqlConn) exec(ctx context.Context, query string, args []driver.Value) (result driver.Result, err error) {
	stmt, err := fc.prepare(ctx, query)
	if err != nil {
		return
	}
	result, err = stmt.(*firebirdsqlStmt).exec(ctx, args)
	if err != nil {
		return
	}
	if fc.isAutocommit && fc.tx.isAutocommit {
		fc.tx.Commit()
	}
	stmt.Close()
	return
}

func (fc *firebirdsqlConn) Exec(query string, args []driver.Value) (result driver.Result, err error) {
	return fc.exec(context.Background(), query, args)
}

func (fc *firebirdsqlConn) query(ctx context.Context, query string, args []driver.Value) (rows driver.Rows, err error) {
	stmt, err := fc.prepare(ctx, query)
	if err != nil {
		return
	}
	rows, err = stmt.(*firebirdsqlStmt).query(ctx, args)
	return
}

func (fc *firebirdsqlConn) Query(query string, args []driver.Value) (rows driver.Rows, err error) {
	return fc.query(context.Background(), query, args)
}

func (fc *firebirdsqlConn) loadTimeZoneId() {
	// TODO: select id, name from rdb$time_zones
	fc.timeZoneIds = map[int]string{
		65535: " GMT",
		65534: " ACT",
		65533: " AET",
		65532: " AGT",
		65531: " ART",
		65530: " AST",
		65529: " Africa / Abidjan",
		65528: " Africa / Accra",
		65527: " Africa / Addis_Ababa",
		65526: " Africa / Algiers",
		65525: " Africa / Asmara",
		65524: " Africa / Asmera",
		65523: " Africa / Bamako",
		65522: " Africa / Bangui",
		65521: " Africa / Banjul",
		65520: " Africa / Bissau",
		65519: " Africa / Blantyre",
		65518: " Africa / Brazzaville",
		65517: " Africa / Bujumbura",
		65516: " Africa / Cairo",
		65515: " Africa / Casablanca",
		65514: " Africa / Ceuta",
		65513: " Africa / Conakry",
		65512: " Africa / Dakar",
		65511: " Africa / Dar_es_Salaam",
		65510: " Africa / Djibouti",
		65509: " Africa / Douala",
		65508: " Africa / El_Aaiun",
		65507: " Africa / Freetown",
		65506: " Africa / Gaborone",
		65505: " Africa / Harare",
		65504: " Africa / Johannesburg",
		65503: " Africa / Juba",
		65502: " Africa / Kampala",
		65501: " Africa / Khartoum",
		65500: " Africa / Kigali",
		65499: " Africa / Kinshasa",
		65498: " Africa / Lagos",
		65497: " Africa / Libreville",
		65496: " Africa / Lome",
		65495: " Africa / Luanda",
		65494: " Africa / Lubumbashi",
		65493: " Africa / Lusaka",
		65492: " Africa / Malabo",
		65491: " Africa / Maputo",
		65490: " Africa / Maseru",
		65489: " Africa / Mbabane",
		65488: " Africa / Mogadishu",
		65487: " Africa / Monrovia",
		65486: " Africa / Nairobi",
		65485: " Africa / Ndjamena",
		65484: " Africa / Niamey",
		65483: " Africa / Nouakchott",
		65482: " Africa / Ouagadougou",
		65481: " Africa/Porto - Novo",
		65480: " Africa / Sao_Tome",
		65479: " Africa / Timbuktu",
		65478: " Africa / Tripoli",
		65477: " Africa / Tunis",
		65476: " Africa / Windhoek",
		65475: " America / Adak",
		65474: " America / Anchorage",
		65473: " America / Anguilla",
		65472: " America / Antigua",
		65471: " America / Araguaina",
		65470: " America / Argentina / Buenos_Aires",
		65469: " America / Argentina / Catamarca",
		65468: " America / Argentina / ComodRivadavia",
		65467: " America / Argentina / Cordoba",
		65466: " America / Argentina / Jujuy",
		65465: " America / Argentina / La_Rioja",
		65464: " America / Argentina / Mendoza",
		65463: " America / Argentina / Rio_Gallegos",
		65462: " America / Argentina / Salta",
		65461: " America / Argentina / San_Juan",
		65460: " America / Argentina / San_Luis",
		65459: " America / Argentina / Tucuman",
		65458: " America / Argentina / Ushuaia",
		65457: " America / Aruba",
		65456: " America / Asuncion",
		65455: " America / Atikokan",
		65454: " America / Atka",
		65453: " America / Bahia",
		65452: " America / Bahia_Banderas",
		65451: " America / Barbados",
		65450: " America / Belem",
		65449: " America / Belize",
		65448: " America/Blanc - Sablon",
		65447: " America / Boa_Vista",
		65446: " America / Bogota",
		65445: " America / Boise",
		65444: " America / Buenos_Aires",
		65443: " America / Cambridge_Bay",
		65442: " America / Campo_Grande",
		65441: " America / Cancun",
		65440: " America / Caracas",
		65439: " America / Catamarca",
		65438: " America / Cayenne",
		65437: " America / Cayman",
		65436: " America / Chicago",
		65435: " America / Chihuahua",
		65434: " America / Coral_Harbour",
		65433: " America / Cordoba",
		65432: " America / Costa_Rica",
		65431: " America / Creston",
		65430: " America / Cuiaba",
		65429: " America / Curacao",
		65428: " America / Danmarkshavn",
		65427: " America / Dawson",
		65426: " America / Dawson_Creek",
		65425: " America / Denver",
		65424: " America / Detroit",
		65423: " America / Dominica",
		65422: " America / Edmonton",
		65421: " America / Eirunepe",
		65420: " America / El_Salvador",
		65419: " America / Ensenada",
		65418: " America / Fort_Nelson",
		65417: " America / Fort_Wayne",
		65416: " America / Fortaleza",
		65415: " America / Glace_Bay",
		65414: " America / Godthab",
		65413: " America / Goose_Bay",
		65412: " America / Grand_Turk",
		65411: " America / Grenada",
		65410: " America / Guadeloupe",
		65409: " America / Guatemala",
		65408: " America / Guayaquil",
		65407: " America / Guyana",
		65406: " America / Halifax",
		65405: " America / Havana",
		65404: " America / Hermosillo",
		65403: " America / Indiana / Indianapolis",
		65402: " America / Indiana / Knox",
		65401: " America / Indiana / Marengo",
		65400: " America / Indiana / Petersburg",
		65399: " America / Indiana / Tell_City",
		65398: " America / Indiana / Vevay",
		65397: " America / Indiana / Vincennes",
		65396: " America / Indiana / Winamac",
		65395: " America / Indianapolis",
		65394: " America / Inuvik",
		65393: " America / Iqaluit",
		65392: " America / Jamaica",
		65391: " America / Jujuy",
		65390: " America / Juneau",
		65389: " America / Kentucky / Louisville",
		65388: " America / Kentucky / Monticello",
		65387: " America / Knox_IN",
		65386: " America / Kralendijk",
		65385: " America / La_Paz",
		65384: " America / Lima",
		65383: " America / Los_Angeles",
		65382: " America / Louisville",
		65381: " America / Lower_Princes",
		65380: " America / Maceio",
		65379: " America / Managua",
		65378: " America / Manaus",
		65377: " America / Marigot",
		65376: " America / Martinique",
		65375: " America / Matamoros",
		65374: " America / Mazatlan",
		65373: " America / Mendoza",
		65372: " America / Menominee",
		65371: " America / Merida",
		65370: " America / Metlakatla",
		65369: " America / Mexico_City",
		65368: " America / Miquelon",
		65367: " America / Moncton",
		65366: " America / Monterrey",
		65365: " America / Montevideo",
		65364: " America / Montreal",
		65363: " America / Montserrat",
		65362: " America / Nassau",
		65361: " America / New_York",
		65360: " America / Nipigon",
		65359: " America / Nome",
		65358: " America / Noronha",
		65357: " America / North_Dakota / Beulah",
		65356: " America / North_Dakota / Center",
		65355: " America / North_Dakota / New_Salem",
		65354: " America / Ojinaga",
		65353: " America / Panama",
		65352: " America / Pangnirtung",
		65351: " America / Paramaribo",
		65350: " America / Phoenix",
		65349: " America/Port - au - Prince",
		65348: " America / Port_of_Spain",
		65347: " America / Porto_Acre",
		65346: " America / Porto_Velho",
		65345: " America / Puerto_Rico",
		65344: " America / Punta_Arenas",
		65343: " America / Rainy_River",
		65342: " America / Rankin_Inlet",
		65341: " America / Recife",
		65340: " America / Regina",
		65339: " America / Resolute",
		65338: " America / Rio_Branco",
		65337: " America / Rosario",
		65336: " America / Santa_Isabel",
		65335: " America / Santarem",
		65334: " America / Santiago",
		65333: " America / Santo_Domingo",
		65332: " America / Sao_Paulo",
		65331: " America / Scoresbysund",
		65330: " America / Shiprock",
		65329: " America / Sitka",
		65328: " America / St_Barthelemy",
		65327: " America / St_Johns",
		65326: " America / St_Kitts",
		65325: " America / St_Lucia",
		65324: " America / St_Thomas",
		65323: " America / St_Vincent",
		65322: " America / Swift_Current",
		65321: " America / Tegucigalpa",
		65320: " America / Thule",
		65319: " America / Thunder_Bay",
		65318: " America / Tijuana",
		65317: " America / Toronto",
		65316: " America / Tortola",
		65315: " America / Vancouver",
		65314: " America / Virgin",
		65313: " America / Whitehorse",
		65312: " America / Winnipeg",
		65311: " America / Yakutat",
		65310: " America / Yellowknife",
		65309: " Antarctica / Casey",
		65308: " Antarctica / Davis",
		65307: " Antarctica / DumontDUrville",
		65306: " Antarctica / Macquarie",
		65305: " Antarctica / Mawson",
		65304: " Antarctica / McMurdo",
		65303: " Antarctica / Palmer",
		65302: " Antarctica / Rothera",
		65301: " Antarctica / South_Pole",
		65300: " Antarctica / Syowa",
		65299: " Antarctica / Troll",
		65298: " Antarctica / Vostok",
		65297: " Arctic / Longyearbyen",
		65296: " Asia / Aden",
		65295: " Asia / Almaty",
		65294: " Asia / Amman",
		65293: " Asia / Anadyr",
		65292: " Asia / Aqtau",
		65291: " Asia / Aqtobe",
		65290: " Asia / Ashgabat",
		65289: " Asia / Ashkhabad",
		65288: " Asia / Atyrau",
		65287: " Asia / Baghdad",
		65286: " Asia / Bahrain",
		65285: " Asia / Baku",
		65284: " Asia / Bangkok",
		65283: " Asia / Barnaul",
		65282: " Asia / Beirut",
		65281: " Asia / Bishkek",
		65280: " Asia / Brunei",
		65279: " Asia / Calcutta",
		65278: " Asia / Chita",
		65277: " Asia / Choibalsan",
		65276: " Asia / Chongqing",
		65275: " Asia / Chungking",
		65274: " Asia / Colombo",
		65273: " Asia / Dacca",
		65272: " Asia / Damascus",
		65271: " Asia / Dhaka",
		65270: " Asia / Dili",
		65269: " Asia / Dubai",
		65268: " Asia / Dushanbe",
		65267: " Asia / Famagusta",
		65266: " Asia / Gaza",
		65265: " Asia / Harbin",
		65264: " Asia / Hebron",
		65263: " Asia / Ho_Chi_Minh",
		65262: " Asia / Hong_Kong",
		65261: " Asia / Hovd",
		65260: " Asia / Irkutsk",
		65259: " Asia / Istanbul",
		65258: " Asia / Jakarta",
		65257: " Asia / Jayapura",
		65256: " Asia / Jerusalem",
		65255: " Asia / Kabul",
		65254: " Asia / Kamchatka",
		65253: " Asia / Karachi",
		65252: " Asia / Kashgar",
		65251: " Asia / Kathmandu",
		65250: " Asia / Katmandu",
		65249: " Asia / Khandyga",
		65248: " Asia / Kolkata",
		65247: " Asia / Krasnoyarsk",
		65246: " Asia / Kuala_Lumpur",
		65245: " Asia / Kuching",
		65244: " Asia / Kuwait",
		65243: " Asia / Macao",
		65242: " Asia / Macau",
		65241: " Asia / Magadan",
		65240: " Asia / Makassar",
		65239: " Asia / Manila",
		65238: " Asia / Muscat",
		65237: " Asia / Nicosia",
		65236: " Asia / Novokuznetsk",
		65235: " Asia / Novosibirsk",
		65234: " Asia / Omsk",
		65233: " Asia / Oral",
		65232: " Asia / Phnom_Penh",
		65231: " Asia / Pontianak",
		65230: " Asia / Pyongyang",
		65229: " Asia / Qatar",
		65228: " Asia / Qyzylorda",
		65227: " Asia / Rangoon",
		65226: " Asia / Riyadh",
		65225: " Asia / Saigon",
		65224: " Asia / Sakhalin",
		65223: " Asia / Samarkand",
		65222: " Asia / Seoul",
		65221: " Asia / Shanghai",
		65220: " Asia / Singapore",
		65219: " Asia / Srednekolymsk",
		65218: " Asia / Taipei",
		65217: " Asia / Tashkent",
		65216: " Asia / Tbilisi",
		65215: " Asia / Tehran",
		65214: " Asia / Tel_Aviv",
		65213: " Asia / Thimbu",
		65212: " Asia / Thimphu",
		65211: " Asia / Tokyo",
		65210: " Asia / Tomsk",
		65209: " Asia / Ujung_Pandang",
		65208: " Asia / Ulaanbaatar",
		65207: " Asia / Ulan_Bator",
		65206: " Asia / Urumqi",
		65205: " Asia/Ust - Nera",
		65204: " Asia / Vientiane",
		65203: " Asia / Vladivostok",
		65202: " Asia / Yakutsk",
		65201: " Asia / Yangon",
		65200: " Asia / Yekaterinburg",
		65199: " Asia / Yerevan",
		65198: " Atlantic / Azores",
		65197: " Atlantic / Bermuda",
		65196: " Atlantic / Canary",
		65195: " Atlantic / Cape_Verde",
		65194: " Atlantic / Faeroe",
		65193: " Atlantic / Faroe",
		65192: " Atlantic / Jan_Mayen",
		65191: " Atlantic / Madeira",
		65190: " Atlantic / Reykjavik",
		65189: " Atlantic / South_Georgia",
		65188: " Atlantic / St_Helena",
		65187: " Atlantic / Stanley",
		65186: " Australia / ACT",
		65185: " Australia / Adelaide",
		65184: " Australia / Brisbane",
		65183: " Australia / Broken_Hill",
		65182: " Australia / Canberra",
		65181: " Australia / Currie",
		65180: " Australia / Darwin",
		65179: " Australia / Eucla",
		65178: " Australia / Hobart",
		65177: " Australia / LHI",
		65176: " Australia / Lindeman",
		65175: " Australia / Lord_Howe",
		65174: " Australia / Melbourne",
		65173: " Australia / NSW",
		65172: " Australia / North",
		65171: " Australia / Perth",
		65170: " Australia / Queensland",
		65169: " Australia / South",
		65168: " Australia / Sydney",
		65167: " Australia / Tasmania",
		65166: " Australia / Victoria",
		65165: " Australia / West",
		65164: " Australia / Yancowinna",
		65163: " BET",
		65162: " BST",
		65161: " Brazil / Acre",
		65160: " Brazil / DeNoronha",
		65159: " Brazil / East",
		65158: " Brazil / West",
		65157: " CAT",
		65156: " CET",
		65155: " CNT",
		65154: " CST",
		65153: " CST6CDT",
		65152: " CTT",
		65151: " Canada / Atlantic",
		65150: " Canada / Central",
		65149: " Canada/East - Saskatchewan",
		65148: " Canada / Eastern",
		65147: " Canada / Mountain",
		65146: " Canada / Newfoundland",
		65145: " Canada / Pacific",
		65144: " Canada / Saskatchewan",
		65143: " Canada / Yukon",
		65142: " Chile / Continental",
		65141: " Chile / EasterIsland",
		65140: " Cuba",
		65139: " EAT",
		65138: " ECT",
		65137: " EET",
		65136: " EST",
		65135: " EST5EDT",
		65134: " Egypt",
		65133: " Eire",
		65132: " Etc / GMT",
		65131: " Etc/GMT + 0",
		65130: " Etc/GMT + 1",
		65129: " Etc/GMT + 10",
		65128: " Etc/GMT + 11",
		65127: " Etc/GMT + 12",
		65126: " Etc/GMT + 2",
		65125: " Etc/GMT + 3",
		65124: " Etc/GMT + 4",
		65123: " Etc/GMT + 5",
		65122: " Etc/GMT + 6",
		65121: " Etc/GMT + 7",
		65120: " Etc/GMT + 8",
		65119: " Etc/GMT + 9",
		65118: " Etc/GMT - 0",
		65117: " Etc/GMT - 1",
		65116: " Etc/GMT - 10",
		65115: " Etc/GMT - 11",
		65114: " Etc/GMT - 12",
		65113: " Etc/GMT - 13",
		65112: " Etc/GMT - 14",
		65111: " Etc/GMT - 2",
		65110: " Etc/GMT - 3",
		65109: " Etc/GMT - 4",
		65108: " Etc/GMT - 5",
		65107: " Etc/GMT - 6",
		65106: " Etc/GMT - 7",
		65105: " Etc/GMT - 8",
		65104: " Etc/GMT - 9",
		65103: " Etc / GMT0",
		65102: " Etc / Greenwich",
		65101: " Etc / UCT",
		65100: " Etc / UTC",
		65099: " Etc / Universal",
		65098: " Etc / Zulu",
		65097: " Europe / Amsterdam",
		65096: " Europe / Andorra",
		65095: " Europe / Astrakhan",
		65094: " Europe / Athens",
		65093: " Europe / Belfast",
		65092: " Europe / Belgrade",
		65091: " Europe / Berlin",
		65090: " Europe / Bratislava",
		65089: " Europe / Brussels",
		65088: " Europe / Bucharest",
		65087: " Europe / Budapest",
		65086: " Europe / Busingen",
		65085: " Europe / Chisinau",
		65084: " Europe / Copenhagen",
		65083: " Europe / Dublin",
		65082: " Europe / Gibraltar",
		65081: " Europe / Guernsey",
		65080: " Europe / Helsinki",
		65079: " Europe / Isle_of_Man",
		65078: " Europe / Istanbul",
		65077: " Europe / Jersey",
		65076: " Europe / Kaliningrad",
		65075: " Europe / Kiev",
		65074: " Europe / Kirov",
		65073: " Europe / Lisbon",
		65072: " Europe / Ljubljana",
		65071: " Europe / London",
		65070: " Europe / Luxembourg",
		65069: " Europe / Madrid",
		65068: " Europe / Malta",
		65067: " Europe / Mariehamn",
		65066: " Europe / Minsk",
		65065: " Europe / Monaco",
		65064: " Europe / Moscow",
		65063: " Europe / Nicosia",
		65062: " Europe / Oslo",
		65061: " Europe / Paris",
		65060: " Europe / Podgorica",
		65059: " Europe / Prague",
		65058: " Europe / Riga",
		65057: " Europe / Rome",
		65056: " Europe / Samara",
		65055: " Europe / San_Marino",
		65054: " Europe / Sarajevo",
		65053: " Europe / Saratov",
		65052: " Europe / Simferopol",
		65051: " Europe / Skopje",
		65050: " Europe / Sofia",
		65049: " Europe / Stockholm",
		65048: " Europe / Tallinn",
		65047: " Europe / Tirane",
		65046: " Europe / Tiraspol",
		65045: " Europe / Ulyanovsk",
		65044: " Europe / Uzhgorod",
		65043: " Europe / Vaduz",
		65042: " Europe / Vatican",
		65041: " Europe / Vienna",
		65040: " Europe / Vilnius",
		65039: " Europe / Volgograd",
		65038: " Europe / Warsaw",
		65037: " Europe / Zagreb",
		65036: " Europe / Zaporozhye",
		65035: " Europe / Zurich",
		65034: " Factory",
		65033: " GB",
		65032: " GB - Eire",
		65031: " GMT + 0",
		65030: " GMT - 0",
		65029: " GMT0",
		65028: " Greenwich",
		65027: " HST",
		65026: " Hongkong",
		65025: " IET",
		65024: " IST",
		65023: " Iceland",
		65022: " Indian / Antananarivo",
		65021: " Indian / Chagos",
		65020: " Indian / Christmas",
		65019: " Indian / Cocos",
		65018: " Indian / Comoro",
		65017: " Indian / Kerguelen",
		65016: " Indian / Mahe",
		65015: " Indian / Maldives",
		65014: " Indian / Mauritius",
		65013: " Indian / Mayotte",
		65012: " Indian / Reunion",
		65011: " Iran",
		65010: " Israel",
		65009: " JST",
		65008: " Jamaica",
		65007: " Japan",
		65006: " Kwajalein",
		65005: " Libya",
		65004: " MET",
		65003: " MIT",
		65002: " MST",
		65001: " MST7MDT",
		65000: " Mexico / BajaNorte",
		64999: " Mexico / BajaSur",
		64998: " Mexico / General",
		64997: " NET",
		64996: " NST",
		64995: " NZ",
		64994: " NZ - CHAT",
		64993: " Navajo",
		64992: " PLT",
		64991: " PNT",
		64990: " PRC",
		64989: " PRT",
		64988: " PST",
		64987: " PST8PDT",
		64986: " Pacific / Apia",
		64985: " Pacific / Auckland",
		64984: " Pacific / Bougainville",
		64983: " Pacific / Chatham",
		64982: " Pacific / Chuuk",
		64981: " Pacific / Easter",
		64980: " Pacific / Efate",
		64979: " Pacific / Enderbury",
		64978: " Pacific / Fakaofo",
		64977: " Pacific / Fiji",
		64976: " Pacific / Funafuti",
		64975: " Pacific / Galapagos",
		64974: " Pacific / Gambier",
		64973: " Pacific / Guadalcanal",
		64972: " Pacific / Guam",
		64971: " Pacific / Honolulu",
		64970: " Pacific / Johnston",
		64969: " Pacific / Kiritimati",
		64968: " Pacific / Kosrae",
		64967: " Pacific / Kwajalein",
		64966: " Pacific / Majuro",
		64965: " Pacific / Marquesas",
		64964: " Pacific / Midway",
		64963: " Pacific / Nauru",
		64962: " Pacific / Niue",
		64961: " Pacific / Norfolk",
		64960: " Pacific / Noumea",
		64959: " Pacific / Pago_Pago",
		64958: " Pacific / Palau",
		64957: " Pacific / Pitcairn",
		64956: " Pacific / Pohnpei",
		64955: " Pacific / Ponape",
		64954: " Pacific / Port_Moresby",
		64953: " Pacific / Rarotonga",
		64952: " Pacific / Saipan",
		64951: " Pacific / Samoa",
		64950: " Pacific / Tahiti",
		64949: " Pacific / Tarawa",
		64948: " Pacific / Tongatapu",
		64947: " Pacific / Truk",
		64946: " Pacific / Wake",
		64945: " Pacific / Wallis",
		64944: " Pacific / Yap",
		64943: " Poland",
		64942: " Portugal",
		64941: " ROC",
		64940: " ROK",
		64939: " SST",
		64938: " Singapore",
		64937: " SystemV / AST4",
		64936: " SystemV / AST4ADT",
		64935: " SystemV / CST6",
		64934: " SystemV / CST6CDT",
		64933: " SystemV / EST5",
		64932: " SystemV / EST5EDT",
		64931: " SystemV / HST10",
		64930: " SystemV / MST7",
		64929: " SystemV / MST7MDT",
		64928: " SystemV / PST8",
		64927: " SystemV / PST8PDT",
		64926: " SystemV / YST9",
		64925: " SystemV / YST9YDT",
		64924: " Turkey",
		64923: " UCT",
		64922: " US / Alaska",
		64921: " US / Aleutian",
		64920: " US / Arizona",
		64919: " US / Central",
		64918: " US/East - Indiana",
		64917: " US / Eastern",
		64916: " US / Hawaii",
		64915: " US/Indiana - Starke",
		64914: " US / Michigan",
		64913: " US / Mountain",
		64912: " US / Pacific",
		64911: " US/Pacific - New",
		64910: " US / Samoa",
		64909: " UTC",
		64908: " Universal",
		64907: " VST",
		64906: " W - SU",
		64905: " WET",
		64904: " Zulu",
	}
}

func newFirebirdsqlConn(dsn string) (fc *firebirdsqlConn, err error) {
	addr, dbName, user, password, options, err := parseDSN(dsn)

	wp, err := newWireProtocol(addr)
	if err != nil {
		return
	}

	column_name_to_lower := convertToBool(options["column_name_to_lower"], false)

	clientPublic, clientSecret := getClientSeed()

	wp.opConnect(dbName, user, password, options, clientPublic)
	err = wp.opAccept(user, password, options, clientPublic, clientSecret)
	if err != nil {
		return
	}
	wp.opAttach(dbName, user, password, options["role"])
	wp.dbHandle, _, _, err = wp.opResponse()
	if err != nil {
		return
	}

	fc = new(firebirdsqlConn)
	fc.wp = wp
	fc.addr = addr
	fc.dbName = dbName
	fc.user = user
	fc.password = password
	fc.columnNameToLower = column_name_to_lower
	fc.isAutocommit = true
	fc.tx, err = newFirebirdsqlTx(fc, ISOLATION_LEVEL_READ_COMMITED, fc.isAutocommit)
	fc.clientPublic = clientPublic
	fc.clientSecret = clientSecret

	fc.loadTimeZoneId()

	return fc, err
}

func createFirebirdsqlConn(dsn string) (fc *firebirdsqlConn, err error) {
	// Create Database
	addr, dbName, user, password, options, err := parseDSN(dsn)

	wp, err := newWireProtocol(addr)
	if err != nil {
		return
	}
	column_name_to_lower := convertToBool(options["column_name_to_lower"], false)

	clientPublic, clientSecret := getClientSeed()

	wp.opConnect(dbName, user, password, options, clientPublic)
	err = wp.opAccept(user, password, options, clientPublic, clientSecret)
	if err != nil {
		return
	}
	wp.opCreate(dbName, user, password, options["role"])
	wp.dbHandle, _, _, err = wp.opResponse()
	if err != nil {
		return
	}

	fc = new(firebirdsqlConn)
	fc.wp = wp
	fc.addr = addr
	fc.dbName = dbName
	fc.user = user
	fc.password = password
	fc.columnNameToLower = column_name_to_lower
	fc.isAutocommit = true
	fc.tx, err = newFirebirdsqlTx(fc, ISOLATION_LEVEL_READ_COMMITED, fc.isAutocommit)
	fc.clientPublic = clientPublic
	fc.clientSecret = clientSecret

	fc.loadTimeZoneId()

	return fc, err
}
