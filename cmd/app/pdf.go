package main

import (
	"fmt"
	"github.com/jung-kurt/gofpdf"
	"os/exec"
)

const (
	robotoCondensed = "RobotoCondensed"
	titleFontSize   = 16
	datefontsize    = 14
	regularFontSize = 12
)

func (x *App) savePDF(fileName string) {
	pdf := gofpdf.New("P", "mm", "A4", "fonts")
	pdf.AddFont(robotoCondensed, "", "RobotoCondensed-Regular.json")
	pdf.AddFont(robotoCondensed, "B", "RobotoCondensed-Bold.json")

	tr := pdf.UnicodeTranslatorFromDescriptor("cp1251")
	//pageWidth, pageHeight := pdf.GetPageSize()

	print := func(w, h float64, align, str string) {
		pdf.CellFormat(w, h, str, "", 0, align, false, 0, "")
	}

	println := func(w, h float64, align, str string) {
		pdf.CellFormat(w, h, str, "", 1, align, false, 0, "")
	}

	pdf.SetLineWidth(0.3)

	for nRecord, a := range x.records {

		if nRecord%4 == 0 {
			pdf.AddPage()
		}

		pdf.SetFont(robotoCondensed, "", titleFontSize)
		str := tr(fmt.Sprintf("Акт визуального и измерительного контроля № %d", a.Order))
		println(0, 10, "C", str)
		pdf.SetFont(robotoCondensed, "B", datefontsize)
		println(0, 8, "R", tr(fmt.Sprintf("от %s", a.Time.Format("02.01.2006"))))

		pdf.SetFont(robotoCondensed, "", regularFontSize)
		str = tr("В соответствии с заявкой ОТКиИ согласно операционным картам  00226247.60103.16125,")
		println(pdf.GetStringWidth(str)+10, 5, "R", str)

		println(0, 5, "L", tr("00226247.60103.16126, 00226247.60103.16127 выполнен визуальный и измерительный контроль (ВиК)"))

		str = tr(fmt.Sprintf("магнитопровода, изделие КЭГ 9721 ИБЯЛ.304135.00%d, партия", a.DocCode))
		print(pdf.GetStringWidth(str)+1, 5, "L", str)

		pdf.SetFont(robotoCondensed, "BU", regularFontSize)
		str = tr(a.RouteSheet)
		print(pdf.GetStringWidth(str), 5, "L", str)

		pdf.SetFont(robotoCondensed, "", regularFontSize)
		println(0, 5, "L", tr(fmt.Sprintf(", в количестве %d штук.", a.ProductsCount)))

		pdf.SetFont(robotoCondensed, "B", regularFontSize)

		pdf.Ln(3)

		str = tr("Заключение по результатам ВиК: ")
		print(pdf.GetStringWidth(str)+10, 5, "R", str)

		pdf.SetFont(robotoCondensed, "", regularFontSize)
		println(0, 5, "L", tr("дефектов не выявлено, объекты контроля соответсвуют"))
		println(0, 5, "L", tr(fmt.Sprintf("требованиям ОСТ 4ГО.070.015, РД 03-606-03 и ИБЯЛ.30413500%d.", a.DocCode)))

		pdf.Ln(5)

		pdf.SetFont(robotoCondensed, "B", regularFontSize)
		str = tr("Контроль выполнил:")
		print(pdf.GetStringWidth(str)+1, 5, "L", str)

		pdf.SetFont(robotoCondensed, "", regularFontSize)
		str = tr("Филимоненков П.А.")
		print(pdf.GetStringWidth(str)+30, 5, "L", str)

		pdf.SetFont(robotoCondensed, "B", regularFontSize)
		str = tr("Начальник ОТКиИ:")
		print(pdf.GetStringWidth(str)+1, 5, "L", str)

		pdf.SetFont(robotoCondensed, "", regularFontSize)
		str = tr("Лемешев В.Л.")
		print(pdf.GetStringWidth(str), 5, "L", str)

		if nRecord%4 != 3 {
			pdf.Ln(12)
			pdf.MoveTo(pdf.GetX(), pdf.GetY())
			pdf.LineTo(pdf.GetX()+190, pdf.GetY())
			pdf.DrawPath("D")
			pdf.Ln(5)
		}
	}
	if err := pdf.OutputFileAndClose(fileName); err != nil {
		panic(err)
	}
	if err := exec.Command("explorer.exe", fileName).Start(); err != nil {
		panic(err)
	}
}
