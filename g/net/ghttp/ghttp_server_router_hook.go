// Copyright 2018 gf Author(https://gitee.com/johng/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://gitee.com/johng/gf.
// 事件回调(中间件)路由控制.

package ghttp

import (
    "strings"
    "container/list"
    "gitee.com/johng/gf/g/util/gregex"
    "gitee.com/johng/gf/g/container/gset"
    "fmt"
)

// 绑定指定的hook回调函数, pattern参数同BindHandler，支持命名路由；hook参数的值由ghttp server设定，参数不区分大小写
func (s *Server)BindHookHandler(pattern string, hook string, handler HandlerFunc) error {
    return s.setHandler(pattern, &handlerItem{
        ctype : nil,
        fname : "",
        faddr : handler,
    }, hook)
    return nil
}

// 通过map批量绑定回调函数
func (s *Server)BindHookHandlerByMap(pattern string, hookmap map[string]HandlerFunc) error {
    for k, v := range hookmap {
        if err := s.BindHookHandler(pattern, k, v); err != nil {
            return err
        }
    }
    return nil
}

// 事件回调处理，内部使用了缓存处理.
// 并按照指定hook回调函数的优先级及注册顺序进行调用
func (s *Server) callHookHandler(hook string, r *Request) {
    hookItems := s.getHookHandlerWithCache(hook, r)
    if len(hookItems) > 0 {
        oldRouterVars := r.routerVars
        for _, item := range hookItems {
            r.routerVars = make(map[string][]string)
            for k, v := range item.values {
                r.routerVars[k] = v
            }
            r.Router = item.handler.router
            item.handler.faddr(r)
        }
        r.routerVars   = oldRouterVars
    }
}

// 查询请求处理方法.
// 内部带锁机制，可以并发读，但是不能并发写；并且有缓存机制，按照Host、Method、Path进行缓存.
func (s *Server) getHookHandlerWithCache(hook string, r *Request) []*handlerParsedItem {
    cacheItems := ([]*handlerParsedItem)(nil)
    cacheKey   := s.hookHandlerKey(hook, r.Method, r.URL.Path, r.GetHost())
    if v := s.hooksCache.Get(cacheKey); v == nil {
        cacheItems = s.searchHookHandler(r.Method, r.URL.Path, r.GetHost(), hook)
        if cacheItems != nil {
            s.hooksCache.Set(cacheKey, cacheItems, 0)
        }
    } else {
        cacheItems = v.([]*handlerParsedItem)
    }
    return cacheItems
}

// 事件方法检索
func (s *Server) searchHookHandler(method, path, domain, hook string) []*handlerParsedItem {
    // 遍历检索的域名列表
    domains := []string{ gDEFAULT_DOMAIN }
    if !strings.EqualFold(gDEFAULT_DOMAIN, domain) {
        domains = append(domains, domain)
    }
    // URL.Path层级拆分
    array := ([]string)(nil)
    if strings.EqualFold("/", path) {
        array = []string{"/"}
    } else {
        array = strings.Split(path[1:], "/")
    }
    parsedItems := make([]*handlerParsedItem, 0)
    for _, domain := range domains {
        p, ok := s.hooksTree[domain]
        if !ok {
            continue
        }
        p, ok  = p.(map[string]interface{})[hook]
        if !ok {
            continue
        }
        // 多层链表(每个节点都有一个*list链表)的目的是当叶子节点未有任何规则匹配时，让父级模糊匹配规则继续处理
        lists := make([]*list.List, 0)
        for k, v := range array {
            if _, ok := p.(map[string]interface{})["*list"]; ok {
                lists = append(lists, p.(map[string]interface{})["*list"].(*list.List))
            }
            if _, ok := p.(map[string]interface{})[v]; ok {
                p = p.(map[string]interface{})[v]
                if k == len(array) - 1 {
                    if _, ok := p.(map[string]interface{})["*list"]; ok {
                        lists = append(lists, p.(map[string]interface{})["*list"].(*list.List))
                        break
                    }
                }
            } else {
                if _, ok := p.(map[string]interface{})["*fuzz"]; ok {
                    p = p.(map[string]interface{})["*fuzz"]
                }
            }
            // 如果是叶子节点，同时判断当前层级的"*fuzz"键名，解决例如：/user/*action 匹配 /user 的规则
            if k == len(array) - 1 {
                if _, ok := p.(map[string]interface{})["*fuzz"]; ok {
                    p = p.(map[string]interface{})["*fuzz"]
                }
                if _, ok := p.(map[string]interface{})["*list"]; ok {
                    lists = append(lists, p.(map[string]interface{})["*list"].(*list.List))
                }
            }
        }

        // 多层链表遍历检索，从数组末尾的链表开始遍历，末尾的深度高优先级也高
        pushedSet := gset.NewStringSet()
        for i := len(lists) - 1; i >= 0; i-- {
            for e := lists[i].Front(); e != nil; e = e.Next() {
                handler := e.Value.(*handlerItem)
                // 动态匹配规则带有gDEFAULT_METHOD的情况，不会像静态规则那样直接解析为所有的HTTP METHOD存储
                if strings.EqualFold(handler.router.Method, gDEFAULT_METHOD) || strings.EqualFold(handler.router.Method, method) {
                    // 注意当不带任何动态路由规则时，len(match) == 1
                    if match, err := gregex.MatchString(handler.router.RegRule, path); err == nil && len(match) > 0 {
                        parsedItem := &handlerParsedItem{handler, nil}
                        // 如果需要query匹配，那么需要重新正则解析URL
                        if len(handler.router.RegNames) > 0 {
                            if len(match) > len(handler.router.RegNames) {
                                parsedItem.values = make(map[string][]string)
                                // 如果存在存在同名路由参数名称，那么执行数组追加
                                for i, name := range handler.router.RegNames {
                                    if _, ok := parsedItem.values[name]; ok {
                                        parsedItem.values[name] = append(parsedItem.values[name], match[i + 1])
                                    } else {
                                        parsedItem.values[name] = []string{match[i + 1]}
                                    }
                                }
                            }
                        }
                        address := fmt.Sprintf("%p", handler)
                        if !pushedSet.Contains(address) {
                            parsedItems = append(parsedItems, parsedItem)
                            pushedSet.Add(address)
                        }
                    }
                }
            }
        }
        return parsedItems
    }
    return nil
}

// 生成hook key
func (s *Server) hookHandlerKey(hook, method, path, domain string) string {
    return hook + "@" + s.serveHandlerKey(method, path, domain)
}

