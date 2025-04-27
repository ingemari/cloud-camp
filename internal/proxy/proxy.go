//3. Прокси proxy/proxy.go
//Используем httputil.NewSingleHostReverseProxy.
//
//Перенаправляем запросы на выбранный сервер.
//
//Обработка ошибок при недоступности бэкенда.

package proxy

import (
	"net/http/httputil"
	"net/url"
)

func NewReverseProxy(target string) (*httputil.ReverseProxy, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	return httputil.NewSingleHostReverseProxy(u), nil
}
