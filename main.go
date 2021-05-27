package main

import (
	"container/list"
	"strings"
	"io/ioutil"
	"flag"
	"fmt"
	"log"
	"os"
	"image"
	"image/png"
	"image/jpeg"
	"runtime"
	"os/exec"
	"strconv"
	"golang.org/x/image/bmp"
)

func listDirectoryRecursive(src string) (l_images *list.List) {
	l_img := list.New()
	archivos, err := ioutil.ReadDir(src)
	if err != nil {
		log.Fatal(err)
	}
	for _, archivo := range archivos {
		if archivo.IsDir() {
			l_img.PushBackList(listDirectoryRecursive(src + "/" + archivo.Name()))
		} else {
			split_r := strings.Split(archivo.Name(), ".")
			extension := strings.ToLower(split_r[len(split_r)-1])
			if extension == "png" || extension == "jpg" || extension == "jpeg" {
				l_img.PushBack(src + "/" + archivo.Name())
			}
		}
	}
	return l_img
}

func vipsExec(src string, dst string, width int, height int) {
	var args = []string{
		"-s", strconv.Itoa(width) + "x" + strconv.Itoa(height) +"!",
		"-o", dst,
		src,
	}
	if runtime.GOOS == "linux" {
		err := exec.Command("vipsthumbnail", args...).Run()
		if err != nil {
			log.Fatal("Error")
		}
	}
}

func createFile(p string) *os.File{
    f, err := os.Create(p)
    if err != nil {
        panic(err)
    }
    return f
}


func generateKittiLabels(clase string, width int, height int, fileName string){
	f := createFile("imagesLabelsKitti/"+strings.Replace(fileName, ".png", ".txt", 1))
    defer f.Close()

	xmin,ymin,xmax,ymax := 0,0,0,0

	if clase == "casco" {
		xmin = width/4
		ymin = 0
		ymax = height/5
		xmax = (width/4)*3
	}else{
		xmin = width/10
		ymin = height/10 + height/20 
		ymax = 5*(height/10)
		xmax = (width/10)*9
	}

	fmt.Fprintln(f, clase, "0.00", "0","0.00", xmin, ymin, xmax, ymax, "0.00", "0.00", "0.00", "0.00", "0.00", "0.00", "0.00")
}

func convertImage(nameFile string, format_dst string){

	imageinput, err := os.Open(nameFile)
	if err != nil {
		log.Fatal(err)
	}
	defer imageinput.Close()

	filenameSplit := strings.Split(nameFile, ".")
	format := filenameSplit[len(filenameSplit)-1]

	if (strings.ToLower(format) != strings.ToLower(format_dst)){
		var src image.Image
		switch strings.ToLower(format) {
		case "png":
			src, err = png.Decode(imageinput)
		case "jpg", "jpeg":
			src, err = jpeg.Decode(imageinput)
		case "bmp":
			src, err = bmp.Decode(imageinput)
		default:
			fmt.Println("The " + format + " we don't support to convert")
			os.Exit(1)
		}
	
		if err != nil {
			log.Fatal(err)
		}
	
		outfile, err := os.Create(filenameSplit[0]+"."+strings.ToLower(format_dst))
		if err != nil {
			log.Fatal(err)
		}
		defer outfile.Close()
	
		switch strings.ToLower(format_dst) {
		case "png":
			err = png.Encode(outfile, src)
		case "jpg", "jpeg":
			err = jpeg.Encode(outfile, src, nil)
		case "bmp":
			err = bmp.Encode(outfile, src)
		default:
			fmt.Println("The " + format + " we don't support to convert")
			os.Exit(1)
		}
		if err != nil {
			log.Fatal(err)
		}
		
		// Borramos original
		err = os.Remove(nameFile)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func main(){
	directory := flag.String("d", "home", "Directorio de imagenes")
	clase := flag.String("c", "casco", "Clases permitidas, casco, chaleco")
	flag.Parse()

    sumax := 0
    sumay := 0
    sumfotos := 0

	l := listDirectoryRecursive(*directory)
	for e := l.Front(); e != nil; e = e.Next() {
		Name := e.Value.(string)
		reader, err := os.Open(Name) 
		defer reader.Close()

		if err != nil {
			fmt.Println("Error: ", err)
			continue
		 }else {
			im, _, _ := image.DecodeConfig(reader)
			sumax += im.Width
			sumay += im.Height
			sumfotos += 1
		 }
		convertImage(Name,"png")
	}
	l = listDirectoryRecursive(*directory)
	for e := l.Front(); e != nil; e = e.Next() {
		rutaImagen := e.Value.(string)
		imagenName:= strings.Split(rutaImagen, "/")
		vipsExec(rutaImagen,"../imagesLabelsKitti/"+imagenName[len(imagenName)-1], sumax/sumfotos, sumay/sumfotos)
		generateKittiLabels(*clase, sumax/sumfotos, sumay/sumfotos, imagenName[len(imagenName)-1])
	}
}