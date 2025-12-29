package helper

import (
	"log"
	"strings"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

func generatePDF(html, nameFile string) error {
	log.Printf("[PDF] begin to unduh pdf")
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}

	page := wkhtmltopdf.NewPageReader(strings.NewReader(html))
	page.EnableLocalFileAccess.Set(true)

	pdfg.AddPage(page)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Dpi.Set(300)

	err = pdfg.Create()
	if err != nil {
		log.Printf("[PDF] Failed to create pdf reader")
		return err
	}

	log.Printf("[PDF] success create pdf with name %s", nameFile)
	return pdfg.WriteFile(nameFile + ".pdf")
}
