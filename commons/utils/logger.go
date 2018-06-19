/*
 *    This file is part of Disgo-Commons library.
 *
 *    The Disgo-Commons library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The Disgo-Commons library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the Disgo-Commons library.  If not, see <http://www.gnu.org/licenses/>.
 */
package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

var loggerOnce sync.Once

// InitializeLogger
func InitializeLogger() {
	loggerOnce.Do(func() {
		// Setup log.
		formatter := &log.TextFormatter{
			FullTimestamp: true,
			ForceColors:   false,
		}
		log.SetFormatter(formatter)
		log.SetOutput(os.Stdout)
		log.SetLevel(log.InfoLevel)
	})
}

// Info
func Info(args ...interface{}) {
	InitializeLogger()
	log.WithFields(log.Fields{
		"method": GetFuncName(2),
	}).Info(args)
}

// Debug
func Debug(args ...interface{}) {
	InitializeLogger()
	log.WithFields(log.Fields{
		"method": GetFuncName(2),
	}).Debug(args)
}

// Warn
func Warn(args ...interface{}) {
	InitializeLogger()
	log.WithFields(log.Fields{
		"method": GetFuncName(2),
	}).Warn(args)
}

// Error
func Error(args ...interface{}) {
	InitializeLogger()
	log.WithFields(log.Fields{
		"method": GetFuncName(2),
	}).Error(args)
}

// Fatal
func Fatal(args ...interface{}) {
	InitializeLogger()
	log.WithFields(log.Fields{
		"method": GetFuncName(2),
	}).Fatal(args)
}
