/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-09-14 17:11:49
 * @LastEditTime: 2023-07-12 10:52:51
 */
package gocommon

import (
	"testing"

	"github.com/reber0/go-common/pkg"
)

func Test_mylog(t *testing.T) {
	log := pkg.NewLog().IsShowCaller(true).IsToFile(true).Logger()
	log.Info("info")
	log.Error("error")
}
