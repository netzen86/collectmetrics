package utils

import (
	"errors"
	"strconv"
)

func ParseValGag(metricValue interface{}) (float64, error) {
	var mv float64
	var convOk error
	switch metricValue := metricValue.(type) {
	case string:
		mv, convOk = strconv.ParseFloat(metricValue, 64)
		if convOk != nil {
			return 0, errors.New("value wrong type")
		}
	case float64:
		mv = metricValue
	default:
		return 0, errors.New("value wrong type")
	}

	return mv, nil
}

func ParseValCnt(metricValue interface{}) (int64, error) {
	var mv int64
	var convOk error
	switch metricValue := metricValue.(type) {
	case string:
		mv, convOk = strconv.ParseInt(metricValue, 10, 64)
		if convOk != nil {
			return 0, errors.New("value wrong type")
		}
	case int64:
		mv = metricValue

	}
	return mv, nil
}
