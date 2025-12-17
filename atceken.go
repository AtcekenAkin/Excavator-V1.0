package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
)

func main() {

	var targetURL string
	var scrapeHTML bool
	var scrapeLinks bool
	var takeshot bool

	flag.StringVar(&targetURL, "url", "", "Hedef Site URl'si.")
	flag.BoolVar(&scrapeHTML, "html", false, "Web sitenin HTML kodu.")
	flag.BoolVar(&scrapeLinks, "links", false, "Linkleri çek.")
	flag.BoolVar(&takeshot, "screenshot", false, "Ekran görüntüsü al.")
	flag.Parse()

	baseName := createFileName(targetURL) //Her URL için o URL'nin adını dosya adı yapma
	htmlName := baseName + ".html"
	shotName := baseName + ".png"
	linksName := baseName + "_linkler.txt"
	//URL kontrolü

	if targetURL == "" {
		fmt.Println("Hata: Lütfen bir URL yazınız. Kullanım: -url http://medium.com")
		os.Exit(1)
	}

	//Ekran görüntüsü alma
	if takeshot {
		fmt.Println("Ekran görüntüsü alınıyor...")
		err := screenshot(targetURL, shotName)
		if err != nil {
			fmt.Printf("\n[!] Ekran görüntüsü hatası: %v\n", err)
		} else {
			fmt.Println("[TAMAMLANDI]")
		}
	}

	c := colly.NewCollector()

	//Bir istek yapmadan önce  "Ziyaret ediliyor..."	 yazdır
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf(" [%s] Ziyaret ediliyor...\n", r.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Hata:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("<-- [%d] Ziyaret Edilen: %s\n", r.StatusCode, r.Request.URL)
	})

	if scrapeHTML {
		c.OnHTML("html", func(e *colly.HTMLElement) {
			htmlcontent, _ := e.DOM.Html()
			saveHTML(htmlName, htmlcontent)
			log.Printf("Web sitesi içeriği çekildi, HTML kodu '%s' dosyasına yazıldı.", htmlName)
		})
	}

	if scrapeLinks {
		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			link := e.Attr("href")
			appendToFile(linksName, link+"\n")
		})
	}

	err := c.Visit(targetURL)
	if err != nil {
		fmt.Println("URL'yi ziyaret ederken bir hata oluştu:", err)
		os.Exit(1)
	}

}

func createFileName(url string) string {
	r := strings.NewReplacer("http://", "", "https://", "", "/", "_", ":", "", ".", "_")
	return r.Replace(url)
}

func saveHTML(filename string, data string) {
	err := os.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		log.Printf("[!] HTML kaydedilirken hata: %v", err)
	}
}
func appendToFile(filename string, data string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Dosyada hata oluştu:%v", err)
		return
	}
	defer f.Close()
	f.WriteString(data)
}

func screenshot(url string, fileName string) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var buf []byte
	if err := chromedp.Run(ctx, chromedp.Navigate(url), chromedp.CaptureScreenshot(&buf)); err != nil {
		return err
	}

	if err := os.WriteFile(fileName, buf, 0644); err != nil {
		return err
	}
	fmt.Println("Ekran görüntüsü başarıyla kaydedildi:", fileName)
	return nil
}
