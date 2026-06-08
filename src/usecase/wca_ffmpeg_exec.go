package usecase

import (
	"context"
	"os/exec"
)

const (
	wcaLinker64  = "/system/bin/linker64"
	wcaFFmpegBin = "/data/data/net.dinglisch.android.taskerm/files/WCA/ffmpeg-wca.bin"
)

func wcaFFmpegCmd(args ...string) *exec.Cmd {
	full := append([]string{wcaFFmpegBin, "ffmpeg"}, args...)
	return exec.Command(wcaLinker64, full...)
}

func wcaFFmpegCmdCtx(ctx context.Context, args ...string) *exec.Cmd {
	full := append([]string{wcaFFmpegBin, "ffmpeg"}, args...)
	return exec.CommandContext(ctx, wcaLinker64, full...)
}
