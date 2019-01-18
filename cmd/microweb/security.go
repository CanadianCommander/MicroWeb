package main

// #include <pwd.h>
// #include <unistd.h>
// #include <stdlib.h>
import "C"

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	mwsettings "github.com/CanadianCommander/MicroWeb/pkg/mwSettings"
)

//AddSecuritySettingDecoders adds setting decoders for security functions
func AddSecuritySettingDecoders() {
	mwsettings.AddSettingDecoder(mwsettings.NewBasicDecoder("security/user"))
	mwsettings.AddSettingDecoder(mwsettings.NewBasicDecoder("security/strict"))
}

/*
EmitSecurityWarning emits a security warning to the log is security settings are missing from the
configuration file.
*/
func EmitSecurityWarning() {
	if !mwsettings.HasSetting("security/user") || !mwsettings.HasSetting("security/strict") {
		logger.LogWarning("Security section missing from configuration!!!! you should probably fix that!")
	}
}

//AbortIfStrict aborts the program immediately if strict security mode is enabled
func AbortIfStrict() {
	if mwsettings.HasSetting("security/strict") && mwsettings.GetSettingBool("security/strict") {
		logger.LogError("Security violation and strict mode enabled! ABORTING!")
		os.Exit(1)
	}
}

/*
DropRootPrivilege drops this programs privilege from that of root to the user named in
the configuration file under security/user
*/
func DropRootPrivilege() {
	// check that we are root
	if syscall.Getuid() == 0 && mwsettings.HasSetting("security/user") {
		uname := mwsettings.GetSettingString("security/user")
		Cuname := C.CString(uname)
		defer C.free(unsafe.Pointer(Cuname))
		userStruct := C.getpwnam(Cuname)

		if userStruct != nil {
			/*
				use low level C functions instead of syscall.Setgid / syscall.Setuid.
				Currently these functions are not implemented as of golang version: go1.10.4 linux/amd64
			*/
			res, errno := C.setgid(userStruct.pw_gid)
			if res != 0 {
				logger.LogError("Failed to set gid with errno: %d", errno)
				AbortIfStrict()
			}

			res, errno = C.setuid(userStruct.pw_uid)
			if res != 0 {
				logger.LogError("Failed to set gid with errno: %d", errno)
				AbortIfStrict()
			}
			logger.LogInfo("ROOT Privieges dropped to those of [%s]", uname)
		} else {
			logger.LogError("Could not drop privilege to specified user: [%s]", uname)
			AbortIfStrict()
		}
	}
}
