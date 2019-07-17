// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_logrus

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/grpclog"
)

// ReplaceGrpcLogger sets the given logrus.Logger as a gRPC-level logger.
// This should be called *before* any other initialization, preferably from init() functions.
func ReplaceGrpcLogger(logger *logrus.Entry) {
	grpclog.SetLoggerV2(newWrapper(logger.WithField("system", SystemField)))
}

func newWrapper(le *logrus.Entry) *lwrapper {
	return &lwrapper{le: le}
}

type lwrapper struct {
	le *logrus.Entry
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print.
func (w *lwrapper) Info(args ...interface{}) {w.le.Info(args...)}

// Infoln logs to INFO log. Arguments are handled in the manner of fmt.Println.
func (w *lwrapper) Infoln(args ...interface{}) {w.le.Infoln(args...)}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func (w *lwrapper) Infof(format string, args ...interface{}) {w.le.Infof(format, args...)}

// Warning logs to WARNING log. Arguments are handled in the manner of fmt.Print.
func (w *lwrapper) Warning(args ...interface{}) {w.le.Warning(args...)}

// Warningln logs to WARNING log. Arguments are handled in the manner of fmt.Println.
func (w *lwrapper) Warningln(args ...interface{}) {w.le.Warningln(args...)}

// Warningf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func (w *lwrapper) Warningf(format string, args ...interface{}) {w.le.Warningf(format, args...)}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print.
func (w *lwrapper) Error(args ...interface{}) {w.le.Error(args...)}

// Errorln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
func (w *lwrapper) Errorln(args ...interface{}) {w.le.Errorln(args...)}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func (w *lwrapper) Errorf(format string, args ...interface{}) {w.le.Errorf(format, args...)}

// Fatal logs to ERROR log. Arguments are handled in the manner of fmt.Print.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (w *lwrapper) Fatal(args ...interface{}) {w.le.Fatal(args...)}

// Fatalln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (w *lwrapper) Fatalln(args ...interface{}) {w.le.Fatalln(args...)}

// Fatalf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (w *lwrapper) Fatalf(format string, args ...interface{}) {w.le.Fatalf(format, args...)}

// V reports whether verbosity level l is at least the requested verbose level.
func (w *lwrapper) V(l int) bool {return int(w.le.Level) >= l}