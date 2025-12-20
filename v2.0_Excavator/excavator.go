package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func main() {
	//Daha havalı olması için
	banner := `
    ███████╗██╗  ██╗ ██████╗ █████╗ ██╗   ██╗ █████╗ ████████╗ ██████╗ ██████╗ 
    ██╔════╝╚██╗██╔╝██╔════╝██╔══██╗██║   ██║██╔══██╗╚══██╔══╝██╔═══██╗██╔══██╗
    █████╗   ╚███╔╝ ██║     ███████║██║   ██║███████║   ██║   ██║   ██║██████╔╝
    ██╔══╝   ██╔██╗ ██║     ██╔══██║╚██╗ ██╔╝██╔══██║   ██║   ██║   ██║██╔══██╗
    ███████╗██╔╝ ██╗╚██████╗██║  ██║ ╚████╔╝ ██║  ██║   ██║   ╚██████╔╝██║  ██║
    ╚══════╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝  ╚═══╝  ╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝
          [ v2.0 EXCAVATOR WEB SCRAPER ] - WAF Bypass & Intel Crawler
               'by Akın Erman ATÇEKEN'
    `
	fmt.Println(banner)

	var targetURL string
	var scrapeHTML bool
	var scrapeLinks bool
	var takeshot bool

	flag.StringVar(&targetURL, "url", "", "Hedef Site URl'si.")
	flag.BoolVar(&scrapeHTML, "html", false, "Web sitenin HTML kodu.")
	flag.BoolVar(&scrapeLinks, "links", false, "Linkleri çek.")
	flag.BoolVar(&takeshot, "screenshot", false, "Ekran görüntüsü al.")
	flag.Parse()

	//URL kontrolü
	if targetURL == "" {
		fmt.Println("Hata: Lütfen bir URL yazınız. Kullanım: -url http://medium.com")
		os.Exit(1)
	}

	//Oluşturulan içeriklerin klasör içine atılması
	dirName := createFileName(targetURL) //Her URL için o URL'nin adını dosya adı yapma
	baseName := dirName
	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		log.Fatalf("Klasör oluşturulurken hata: %v", err)
	}
	fmt.Printf("[*] Çıktılar %s klasörüne kaydedilecek.\n", dirName)

	htmlName := dirName + "/" + baseName + ".html"
	shotName := dirName + "/" + baseName + ".png"
	linksName := dirName + "/" + baseName + "_linkler.txt"

	//TLS İstemcisi Oluşturma
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_120),
		tls_client.WithTimeout(int((30 * time.Second) / time.Millisecond)),
		tls_client.WithInsecureSkipVerify(),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Fatalf("[!] > TLS istemcisi oluşturulurken bir hata oldu: %v", err)
	}

	//İstek gönderme
	fmt.Printf("[*] Hedefe istek gönderiliyor: %s\n", targetURL)
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		log.Fatalf("[!] > İstek oluşturulurken bir hata oluştu: %v", err)
	}

	//Header ekleme

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("[!] > İstek gönderilirken bir hata oluştu: %v", err)

	}
	defer resp.Body.Close()

	fmt.Printf("[*] Yanıt alındı. Durum Kodu: %d\n", resp.StatusCode)
	bodyBytes, err := io.ReadAll(resp.Body)
	htmlContent := string(bodyBytes)

	if scrapeHTML {
		saveHTML(htmlName, htmlContent)
		fmt.Printf("[*] HTML içeriği %s dosyasına kaydedildi.\n", htmlName)
	}

	if scrapeLinks {
		fmt.Println("[*] Linkler çekiliyor...")
		lines := strings.Split(htmlContent, "\n")
		for _, line := range lines {
			if strings.Contains(line, "href=") {
				start := strings.Index(line, "href=") + 6
				end := strings.Index(line[start:], "\"") + start
				if end > start {
					link := line[start:end]
					appendToFile(linksName, link+"\n")
				}
			}
		}
	}
	fmt.Printf("[*] Linkler %s dosyasına kaydedildi.\n", linksName)

	//Ekran görüntüsü alma
	if takeshot {
		fmt.Println(" [*] Ekran görüntüsü alınıyor...")
		err := screenshot(targetURL, shotName)
		if err != nil {
			fmt.Printf("\n[!] Ekran görüntüsü hatası: %v\n", err)
		} else {
			fmt.Printf("[TAMAMLANDI] Ekran görüntüsü kaydedildi: %s\n", shotName)
		}
	}
}

func createFileName(url string) string {
	r := strings.NewReplacer("http://", "", "https://", "", "/", "_", ":", "", ".", "_", "?", "_", "=", "_", "&", "_", "*", "_", "|", "_", "\"", "_", "<", "_", ">", "_")
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

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.WindowSize(1920, 1080),
		chromedp.Flag("lang", "tr-TR,tr"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx, chromedp.Navigate(url), chromedp.Sleep(5*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, _, contentSize, _, _, _, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(contentSize.Width), int64(contentSize.Height)
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).Do(ctx)
			if err != nil {
				return err
			}

			buf, err = page.CaptureScreenshot().
				WithQuality(90).
				WithFormat(page.CaptureScreenshotFormatPng).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			return err
		}),
	)

	if err != nil {
		return err
	}
	fmt.Println("Ekran görüntüsü başarıyla kaydedildi:", fileName)
	return os.WriteFile(fileName, buf, 0644)
}
