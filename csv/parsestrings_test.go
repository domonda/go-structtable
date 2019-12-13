package csv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ungerik/go-fs"
)

var testRows = map[string][]string{
	"A;\"Line1\nLine2\";B": {
		";", // separator
		"A",
		"Line1\nLine2",
		"B",
	},
	"A;\"Line1\r\nLine2\";B\r\n": {
		";", // separator
		"A",
		"Line1\nLine2",
		"B",
	},
	"A;\"Line1\r\nLine2\";B\r\r\n": {
		";", // separator
		"A",
		"Line1\nLine2",
		"B",
	},
	` Hello ,World ,	!`: {
		",",
		` Hello `,
		`World `,
		`	!`,
	},
	"\n\n\n Hello ,World ,	!\n\n\n": {
		",",
		` Hello `,
		`World `,
		`	!`,
	},
	`" Hello ","World ","	!"`: {
		",",
		` Hello `,
		`World `,
		`	!`,
	},
	`1997,Ford,E350,"Super, luxurious truck"`: {
		",",
		`1997`,
		`Ford`,
		`E350`,
		`Super, luxurious truck`,
	},
	`1997,Ford,E350,"Super, ""luxurious"" truck"`: {
		",",
		`1997`,
		`Ford`,
		`E350`,
		`Super, "luxurious" truck`,
	},
	`1997,""Ford"",E350,"Super, luxurious truck"`: {
		",",
		`1997`,
		`"Ford"`,
		`E350`,
		`Super, luxurious truck`,
	},
	`1997,"""Ford""",E350,"Super, luxurious truck"`: {
		",",
		`1997`,
		`"Ford"`,
		`E350`,
		`Super, luxurious truck`,
	},
	`"1997","""Ford""","E350","Super, luxurious truck"`: {
		",",
		`1997`,
		`"Ford"`,
		`E350`,
		`Super, luxurious truck`,
	},
	// TODO failing:
	// `"1997","Ford,"E350","""Super, luxurious truck"""`: {
	// 	",",
	// 	`1997`,
	// 	`Ford`,
	// 	`E350`,
	// 	`"Super, luxurious truck"`,
	// },
	`05.10.2018;""Heimbau"" Gemeinnützige Bau-, Wohnungs- u. Siedlungsgenossenscha;AT4112xxxxx;BKAUATWWXXX;;;-85,91;EUR;ENTGELT 10/2018 ""Heimbau"" Gemeinnützige Bau-, Wohnu;12000;;0;05.10.2018`: {
		";", // separator
		`05.10.2018`,
		`"Heimbau" Gemeinnützige Bau-, Wohnungs- u. Siedlungsgenossenscha`,
		`AT4112xxxxx`,
		`BKAUATWWXXX`,
		``,
		``,
		`-85,91`,
		`EUR`,
		`ENTGELT 10/2018 "Heimbau" Gemeinnützige Bau-, Wohnu`,
		`12000`,
		``,
		`0`,
		`05.10.2018`,
	},
	`26.06.2018,25.06.2018,Kreditkarte,"-42,87",EUR,"COURSERA inkl. Fremdwährungsentgelt 0,63 Kurs 1,1600378",`: {
		",", // separator
		`26.06.2018`,
		`25.06.2018`,
		`Kreditkarte`,
		`-42,87`,
		`EUR`,
		`COURSERA inkl. Fremdwährungsentgelt 0,63 Kurs 1,1600378`,
		``,
	},
	`"30.12.2018","21:56:09","CET","charlieBAUM DIVERS ET IMPREVU","PayPal Express-Zahlung","Abgeschlossen","EUR","76,80","-2,42","74,38","charliebaum@wanadoo.fr","joerg@saturo.eu","0PE15874WY2156812","isabelle darrigrand, 15 AVENUE EDOUARD VII, INTERPHONE ""LE 4"", BIARRITZ, 64200, Frankreich","Bestätigt","Ready To Drink - 330 ml - Original, Ready To Drink - 330 ml - Strawberry","","0,00","","0,00","","","","","","201812300043437","{""order_id"":198790,""order_number"":""201812300043437"",""order_key"":""wc_order_5c2930bb3e682""}","5","","6.780,42","15 AVENUE EDOUARD VII","INTERPHONE ""LE 4""","BIARRITZ","","64200","Frankreich","0607069536","Ready To Drink - 330 ml - Original","","Sofort","","T0006","","FR","FR","Haben"`: {
		",", // separator
		"30.12.2018",
		"21:56:09",
		"CET",
		"charlieBAUM DIVERS ET IMPREVU",
		"PayPal Express-Zahlung",
		"Abgeschlossen",
		"EUR",
		"76,80",
		"-2,42",
		"74,38",
		"charliebaum@wanadoo.fr",
		"joerg@saturo.eu",
		"0PE15874WY2156812",
		`isabelle darrigrand, 15 AVENUE EDOUARD VII, INTERPHONE "LE 4", BIARRITZ, 64200, Frankreich`,
		"Bestätigt",
		"Ready To Drink - 330 ml - Original, Ready To Drink - 330 ml - Strawberry",
		"",
		"0,00",
		"",
		"0,00",
		"",
		"",
		"",
		"",
		"",
		"201812300043437",
		`{"order_id":198790,"order_number":"201812300043437","order_key":"wc_order_5c2930bb3e682"}`,
		"5",
		"",
		"6.780,42",
		"15 AVENUE EDOUARD VII",
		`INTERPHONE "LE 4"`,
		"BIARRITZ",
		"",
		"64200",
		"Frankreich",
		"0607069536",
		"Ready To Drink - 330 ml - Original",
		"",
		"Sofort",
		"",
		"T0006",
		"",
		"FR",
		"FR",
		"Haben",
	},
	`"15.12.2019","""Heimbau"" Gemeinnützige Bau-, Wohnungs- u. Siedlungsgenossenscha","AT","BKAUATWWXXX","","12000","-8,70","EUR","ENTGELT","xxxxx","","0","15.12.2019","","","","","0-9x9-05","ATx"`: {
		",", // separator
		"15.12.2019",
		"\"Heimbau\" Gemeinnützige Bau-, Wohnungs- u. Siedlungsgenossenscha",
		"AT",
		"BKAUATWWXXX",
		"",
		"12000",
		"-8,70",
		"EUR",
		"ENTGELT",
		"xxxxx",
		"",
		"0",
		"15.12.2019",
		"",
		"",
		"",
		"",
		"0-9x9-05",
		"ATx",
	},
}

func TestParseStrings(t *testing.T) {
	for csvRow, ref := range testRows {
		t.Run(csvRow, func(t *testing.T) {
			refSeparator, refFields := ref[0], ref[1:]
			rows, format, err := ParseStringsDetectFormat([]byte(csvRow), NewFormatDetectionConfig())
			assert.NoError(t, err, "csv.Read")
			assert.NotNil(t, format, "returned Format")
			assert.Equal(t, "UTF-8", format.Encoding, "UTF-8 encoding expected")
			assert.Equalf(t, refSeparator, format.Separator, "'s' separator expected", refSeparator)
			SetRowsWithNonUniformColumnsNil(rows)
			rows = RemoveEmptyRows(rows)
			assert.Len(t, rows, 1, "one CSV row expected")
			if len(rows) == 1 {
				rowFields := rows[0]
				assert.Equal(t, len(refFields), len(rowFields), "parsed CSV row field count")
				for i := range rowFields {
					assert.Equalf(t, refFields[i], rowFields[i], "parsed CSV row field %d", i)
				}
			}
		})
	}

}

func TestParsePriavteStrings(t *testing.T) {
	privateTestDataDir := fs.File("../../TestDocuments/CSV")
	assert.True(t, privateTestDataDir.IsDir(), "privateTestDataDir exists")

	type Expected struct {
		Format *Format
		Rows   [][]string
	}

	testCSV := func(csvFile fs.File) error {
		jsonFile := csvFile.TrimExt() + ".json"
		assert.True(t, jsonFile.Exists())

		var expected Expected
		err := jsonFile.ReadJSON(&expected)
		assert.NoError(t, err, "ReadJSON")

		rows, format, err := FileParseStringsDetectFormat(csvFile, NewFormatDetectionConfig())
		assert.NoError(t, err, "FileParseStringsDetectFormat")
		rows = RemoveEmptyRows(rows)

		assert.Equal(t, expected.Format, format, "detected format")
		assert.Equalf(t, expected.Rows, rows, "rows from %s equal to %s", jsonFile, csvFile)

		return nil
	}

	privateTestDataDir.ListDir(testCSV, "*.csv")
}
