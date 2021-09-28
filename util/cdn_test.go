// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertSignatureCDN(t *testing.T) {
	cdnURL := "https://p26-sign.douyinpic.com/obj/tos-cn-i-0813/0c9111db102d40f89792c5aa18e14581?x-expires=1634058000&x-signature=DJ80mn6D3slXuQZWlwFKkLKDENI%3D&from=4257465056_large"
	got := ConvertSignatureCDN(cdnURL)
	want := "https://p26.douyinpic.com/obj/tos-cn-i-0813/0c9111db102d40f89792c5aa18e14581"
	assert.Equal(t, want, got)
}

func TestIsGIFImage(t *testing.T) {
	gifURL := "https://p6.douyinpic.com/obj/tos-cn-p-0015/a9a12bbd889f41e28fabfcd5669a266e_1632729560"
	got := IsGIFImage(gifURL)
	assert.True(t, got)

	staticURL := "https://p3.douyinpic.com/obj/tos-cn-i-0813/0c9111db102d40f89792c5aa18e14581"
	got = IsGIFImage(staticURL)
	assert.False(t, got)
}
