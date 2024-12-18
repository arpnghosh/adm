package ftpdownload

import (
	"io"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/jlaffaye/ftp"
)

// DownloadFile will download a file from the FTP server.
// this code does not make use of segments
func DownloadFile(rawURL string, parsedURL *url.URL, segment int) {
	host := parsedURL.Host
	user := parsedURL.User.Username()
	pass, _ := parsedURL.User.Password()
	path := parsedURL.Path

	// Connect to FTP server
	c, err := ftp.Dial(host+":21", ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal("Error connecting to FTP server:", err)
	}
	defer c.Quit()

	// Login to the FTP server
	err = c.Login(user, pass)
	if err != nil {
		log.Fatal("Error logging in to FTP server:", err)
	}

	filename := "download_" + path
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal("Error creating the file:", err)
	}
	defer file.Close()

	response, err := c.Retr(path)
	if err != nil {
		log.Fatal("Error retrieving the file:", err)
	}
	defer response.Close()

	// Copy the file data into the created file
	_, err = io.Copy(file, response)
	if err != nil {
		log.Fatal("Error copying data to file:", err)
	}
}
