package gopdf

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"testing"

	"github.com/skirrund/go-pdf/gopdffork"
)

func TestT(t *testing.T) {
	pdf, err := New()
	if err != nil {
		fmt.Println(err)
		return
	}
	pageSize := gopdffork.PageSizeA4Landscape
	pdf.SetMargins(50, 200, 50, 200)
	pdf.AddPageWithOption(gopdffork.PageOption{PageSize: pageSize})
	titleFontSize := 20.0
	pdf.SetFont("song", "", titleFontSize)
	title := "记账凭证"
	x, _ := AlignCenterInPageX(pdf, pageSize, title)
	y := 70.0
	pdf.SetXY(x, y)
	pdf.Text(title)
	y = y + titleFontSize + 5
	contentFontSize := 10.0
	pdf.SetFont("song", "", contentFontSize)
	period := "2025/12/02"
	x, _ = AlignCenterInPageX(pdf, pageSize, period)
	pdf.SetXY(x, y)
	pdf.Text(period)

	// Set the starting Y position for the table
	tableStartY := 120.0
	// Set the left margin for the table
	marginLeft := 50.0
	//核算单位
	hsdw := "核算单位：上海宸汐科技集团有限公司"
	h, _ := pdf.MeasureCellHeightByText(hsdw)
	pdf.SetXY(marginLeft, tableStartY-contentFontSize-2)
	pdf.Text(hsdw)

	pzh := "凭证号：记-50"
	ym := ""
	pzw, _ := pdf.MeasureTextWidth(pzh)
	ymw, _ := pdf.MeasureTextWidth(ym)

	pdf.SetXY(pageSize.W-pzw-ymw-marginLeft-100, tableStartY-h-2)
	pdf.Text(pzh)
	pdf.SetXY(pageSize.W-ymw-marginLeft-50, tableStartY-h-2)
	pdf.Text(ym)
	fj := "附件张数：0"
	pdf.SetXY(pageSize.W-pzw-ymw-marginLeft-100, tableStartY-h*2-10)
	pdf.Text(fj)
	jzr := "记账人："
	y = tableStartY + 34*10 + 40
	pdf.SetXY(marginLeft, y)
	pdf.Text(jzr)

	shr := "审核人："
	x, _ = AlignCenterInPageX(pdf, pageSize, shr)
	pdf.SetXY(x, y)
	pdf.Text(shr)

	zdr := "制单人：王文政"
	zdrw, _ := pdf.MeasureTextWidth(zdr)
	x = pageSize.W - marginLeft - zdrw
	pdf.SetXY(x, y)
	pdf.Text(zdr)

	// Create a new table layout
	table := pdf.NewTableLayout(marginLeft, tableStartY, 34, 9)
	// Add columns to the table
	//742
	table.AddColumn("分录号", 100, "center")
	table.AddColumn("摘要", 242, "left")
	table.AddColumn("科目", 200, "left")
	table.AddColumn("借方金额", 100, "right")
	table.AddColumn("贷方金额", 100, "right")
	table.AddRow([]string{"001", "张二超报销差旅费-小交通-员工报销-B2500834512312312312312312", "3", "5.00", "15.00"})
	table.AddRow([]string{"", "", "", "", ""})
	table.AddRow([]string{"", "", "", "", ""})
	table.AddRow([]string{"", "", "", "", ""})
	table.AddRow([]string{"", "", "", "", ""})
	table.AddRow([]string{"", "", "", "", ""})
	table.AddRow([]string{"", "", "", "", ""})
	table.AddRow([]string{"", "", "", "", ""})
	table.AddRow([]string{"合计", "大写", "", "5.00", "15.00"})
	table.SetTableStyle(gopdffork.CellStyle{
		BorderStyle: gopdffork.BorderStyle{
			Top:    false,
			Left:   false,
			Bottom: false,
			Right:  false,
			Width:  0.5,
		},
		FillColor: gopdffork.RGBColor{R: 255, G: 255, B: 255},
		TextColor: gopdffork.RGBColor{R: 0, G: 0, B: 0},
	})
	// Set the style for table header
	table.SetHeaderStyle(gopdffork.CellStyle{
		BorderStyle: gopdffork.BorderStyle{
			Top:    true,
			Left:   true,
			Bottom: false,
			Right:  true,
			Width:  0.5,
		},
		FillColor: gopdffork.RGBColor{R: 255, G: 255, B: 255},
		TextColor: gopdffork.RGBColor{R: 0, G: 0, B: 0},
	})
	table.SetCellStyle(gopdffork.CellStyle{
		BorderStyle: gopdffork.BorderStyle{
			Right:  true,
			Bottom: true,
			Top:    true,
			Left:   true,
			Width:  0.5,
		},
		FillColor: gopdffork.RGBColor{R: 255, G: 255, B: 255},
		TextColor: gopdffork.RGBColor{R: 0, G: 0, B: 0},
	})
	table.DrawTable()
	f, _ := os.Open("/Users/jerry.shi/Desktop/宸汐健康/voucher_template.pdf")
	Merge(pdf, f)
	pdf.WritePdf("记账凭证_test.pdf")

}

func TestAddKeywords(t *testing.T) {
	ls := []*PDFSearchLocation{
		&PDFSearchLocation{
			Page:     2,
			AddText:  "姓名",
			AbsX:     180,
			AbsY:     587,
			FontSize: 12,
		},
		&PDFSearchLocation{
			Page:     2,
			AddText:  "男",
			AbsX:     430,
			AbsY:     587,
			FontSize: 12,
		},
		&PDFSearchLocation{
			Page:     2,
			AddText:  "31000000000",
			AbsX:     180,
			AbsY:     543,
			FontSize: 12,
		},
		&PDFSearchLocation{
			Page:     2,
			AddText:  "2023",
			AbsX:     160,
			AbsY:     407,
			FontSize: 12,
		},
		&PDFSearchLocation{
			Page:     2,
			AddText:  "02",
			AbsX:     210,
			AbsY:     407,
			FontSize: 12,
		},
		&PDFSearchLocation{
			Page:     2,
			AddText:  "31",
			AbsX:     246,
			AbsY:     407,
			FontSize: 12,
		},
		&PDFSearchLocation{
			Page:     2,
			AddText:  "13321954022",
			AbsX:     180,
			AbsY:     497,
			FontSize: 12,
		},
		&PDFSearchLocation{
			Page:     2,
			AddText:  "方案名称",
			AbsX:     180,
			AbsY:     451,
			FontSize: 12,
			// Image: &Image{
			// 	Url:    "/Users/jerry.shi/Desktop/多少天海报_03.png",
			// 	Width:  100,
			// 	Height: 50,
			// },
		},
	}
	// bs, err := os.ReadFile("/Users/jerry.shi/Desktop/多少天海报_03.png")
	// if err != nil {
	// 	t.Log("<<<<<<<<<<<<<<<", err)
	// }
	// str := base64.StdEncoding.EncodeToString(bs)
	ls[len(ls)-1].Image = nil

	b, err := os.ReadFile("/Users/jerry.shi/Desktop/383_s.pdf")
	if err != nil {
		t.Log("<<<<<<<<<<<<<<<", err)
	}

	b, err = AddKeywordsBytes(ls, b, true)
	// err := AddKeywords(ls, "/Users/jerry.shi/Desktop/240_s.pdf", "/Users/jerry.shi/Desktop/test_s.pdf", true)
	if err != nil {
		t.Error(">>>>>>>>>>>", err)
	}
	err = os.WriteFile("/Users/jerry.shi/Desktop/test_s.pdf", b, fs.ModePerm)
	if err != nil {
		t.Error(">>>>>>>>>>>", err)
	}
	t.Log("end")
}

func TestImg(t *testing.T) {
	resp, err := http.Get("https://wework.qpic.cn/bizmail/CYwp2sODdzz5tItmMicoBfibxibiakMe2bGOSib9hPmKIUWRJiaiaQeBy4kqw/0")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	f, err := os.Create("/Users/jerry.shi/Desktop/testQW.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.Write(bytes)
}
