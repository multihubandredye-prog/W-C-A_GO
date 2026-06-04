package usecase

import (
	"context"
	"os/exec"
)

const (
	wcaLinker64  = "/system/bin/linker64"
	wcaFFmpegBin = "/data/data/net.dinglisch.android.taskerm/files/WCA/ffmpeg-wca.bin"
)

// wcaFFmpegCmd substitui exec.Command("ffmpeg", args...)
func wcaFFmpegCmd(args ...string) *exec.Cmd {
	full := append([]string{wcaFFmpegBin, "ffmpeg"}, args...)
	return exec.Command(wcaLinker64, full...)
}

// wcaFFmpegCmdCtx substitui exec.CommandContext(ctx, "ffmpeg", args...)
func wcaFFmpegCmdCtx(ctx context.Context, args ...string) *exec.Cmd {
	full := append([]string{wcaFFmpegBin, "ffmpeg"}, args...)
	return exec.CommandContext(ctx, wcaLinker64, full...)
}
