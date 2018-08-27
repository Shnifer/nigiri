package main

import (
	"github.com/hajimehoshi/ebiten"
	"io/ioutil"
	"path"
	"image"
	"bytes"
)

func FileTexLoader(pathStr string) TexLoaderF{
	return func(name string) (*ebiten.Image, error) {
		dat, err:=ioutil.ReadFile(path.Join(pathStr, name))
		if err!=nil{
			return nil,err
		}
		buf:=bytes.NewBuffer(dat)
		img, _, err:=image.Decode(buf)
		if err!=nil{
			return nil,err
		}
		ebiImg, err:=ebiten.NewImageFromImage(img, ebiten.FilterDefault)
		if err!=nil{
			return nil,err
		}
		return ebiImg, nil
	}
}