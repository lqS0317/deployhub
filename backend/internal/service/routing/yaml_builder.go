package routing

import (
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

// --- 配置结构体：用于解析 RouteEntry.Config JSONB ---

// ServiceConfig K8s Service 配置
type ServiceConfig struct {
	Type     string            `json:"type"`
	Selector map[string]string `json:"selector"`
	Ports    []ServicePort     `json:"ports"`
}

// ServicePort 端口定义
type ServicePort struct {
	Name       string `json:"name"`
	Port       int    `json:"port"`
	TargetPort int    `json:"targetPort"`
	Protocol   string `json:"protocol"`
}

// IngressConfig K8s Ingress 配置
type IngressConfig struct {
	IngressClassName string            `json:"ingressClassName"`
	TLS              []IngressTLS      `json:"tls"`
	Rules            []IngressRule     `json:"rules"`
	Annotations      map[string]string `json:"annotations"`
}

// IngressTLS TLS 配置
type IngressTLS struct {
	Hosts      []string `json:"hosts"`
	SecretName string   `json:"secretName"`
}

// IngressRule 规则配置
type IngressRule struct {
	Host  string        `json:"host"`
	Paths []IngressPath `json:"paths"`
}

// IngressPath 路径配置
type IngressPath struct {
	Path     string `json:"path"`
	PathType string `json:"pathType"`
	Backend  struct {
		ServiceName string `json:"serviceName"`
		ServicePort int    `json:"servicePort"`
	} `json:"backend"`
}

// IngressRouteConfig Traefik IngressRoute 配置
type IngressRouteConfig struct {
	EntryPoints []string              `json:"entryPoints"`
	Routes      []IngressRouteEntry   `json:"routes"`
	TLS         *IngressRouteTLS      `json:"tls,omitempty"`
	Annotations map[string]string     `json:"annotations,omitempty"`
}

// IngressRouteEntry 路由条目
type IngressRouteEntry struct {
	Match       string                    `json:"match"`
	Kind        string                    `json:"kind"`
	Priority    int                       `json:"priority,omitempty"`
	Services    []IngressRouteService     `json:"services"`
	Middlewares []IngressRouteMiddleware   `json:"middlewares,omitempty"`
}

// IngressRouteMiddleware 中间件引用
type IngressRouteMiddleware struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// IngressRouteService 后端服务
type IngressRouteService struct {
	Name           string `json:"name"`
	Port           int    `json:"port"`
	Namespace      string `json:"namespace,omitempty"`
	Kind           string `json:"kind,omitempty"`
	PassHostHeader *bool  `json:"passHostHeader,omitempty"`
	Scheme         string `json:"scheme,omitempty"`
	NativeLB       bool   `json:"nativeLB,omitempty"`
	Weight         int    `json:"weight,omitempty"`
}

// IngressRouteTLS Traefik TLS 配置
type IngressRouteTLS struct {
	SecretName   string                `json:"secretName,omitempty"`
	CertResolver string                `json:"certResolver,omitempty"`
	Domains      []IngressRouteDomain  `json:"domains,omitempty"`
	Options      *IngressRouteTLSRef   `json:"options,omitempty"`
}

type IngressRouteDomain struct {
	Main string   `json:"main"`
	SANs []string `json:"sans,omitempty"`
}

type IngressRouteTLSRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// ApisixRouteConfig APISIX Route 配置
// ApisixRouteConfig APISIX Route 配置（对应 spec）
type ApisixRouteConfig struct {
	HTTP        []ApisixHTTPRule  `json:"http"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ApisixHTTPRule HTTP 路由规则（对应 spec.http[]）
type ApisixHTTPRule struct {
	Name     string            `json:"name"`
	Priority int               `json:"priority,omitempty"`
	Match    ApisixMatch       `json:"match"`
	Backends []ApisixBackend   `json:"backends"`
	Plugins  []ApisixPlugin    `json:"plugins,omitempty"`
	Timeout  *ApisixTimeout    `json:"timeout,omitempty"`
	PluginConfigName string    `json:"plugin_config_name,omitempty"`
}

// ApisixMatch 请求匹配条件
type ApisixMatch struct {
	Hosts       []string `json:"hosts,omitempty"`
	Paths       []string `json:"paths"`
	Methods     []string `json:"methods,omitempty"`
	RemoteAddrs []string `json:"remoteAddrs,omitempty"`
}

// ApisixBackend 后端服务
type ApisixBackend struct {
	ServiceName        string `json:"serviceName"`
	ServicePort        int    `json:"servicePort"`
	Weight             int    `json:"weight,omitempty"`
	ResolveGranularity string `json:"resolveGranularity,omitempty"`
	Subset             string `json:"subset,omitempty"`
}

// ApisixTimeout 超时配置
type ApisixTimeout struct {
	Connect string `json:"connect,omitempty"`
	Read    string `json:"read,omitempty"`
	Send    string `json:"send,omitempty"`
}

// ApisixPlugin 插件配置
type ApisixPlugin struct {
	Name   string                 `json:"name"`
	Enable bool                   `json:"enable"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// ApisixUpstreamConfig APISIX Upstream 配置（用于 gRPC 等场景）
type ApisixUpstreamConfig struct {
	Scheme        string                    `json:"scheme,omitempty"`
	LoadBalancer  *ApisixLoadBalancer       `json:"loadbalancer,omitempty"`
	Retries       int                       `json:"retries,omitempty"`
	Timeout       *ApisixTimeout            `json:"timeout,omitempty"`
	HealthCheck   map[string]interface{}    `json:"healthCheck,omitempty"`
	PortLevelSettings []ApisixPortLevel     `json:"portLevelSettings,omitempty"`
	Annotations   map[string]string         `json:"annotations,omitempty"`
}

// ApisixLoadBalancer 负载均衡配置
type ApisixLoadBalancer struct {
	Type     string `json:"type"`
	HashOn   string `json:"hashOn,omitempty"`
	Key      string `json:"key,omitempty"`
}

// ApisixPortLevel 端口级别配置（不同端口可以用不同 scheme）
type ApisixPortLevel struct {
	Port   int    `json:"port"`
	Scheme string `json:"scheme"`
}

// BuildYAML 根据资源类型分发到对应的 YAML 构建函数
func BuildYAML(name, namespace, resourceType string, config datatypes.JSON) (string, error) {
	switch resourceType {
	case "service":
		return buildK8sServiceYAML(name, namespace, config)
	case "ingress":
		return buildIngressYAML(name, namespace, config)
	case "ingressroute":
		return buildIngressRouteYAML(name, namespace, config)
	case "apisixroute":
		return buildApisixRouteYAML(name, namespace, config)
	case "apisixupstream":
		return buildApisixUpstreamYAML(name, namespace, config)
	default:
		return "", fmt.Errorf("不支持的资源类型: %s", resourceType)
	}
}

// buildK8sServiceYAML 构建 K8s Service YAML
func buildK8sServiceYAML(name, namespace string, config datatypes.JSON) (string, error) {
	var cfg ServiceConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return "", fmt.Errorf("解析 Service 配置失败: %w", err)
	}

	svcType := corev1.ServiceTypeClusterIP
	switch cfg.Type {
	case "NodePort":
		svcType = corev1.ServiceTypeNodePort
	case "LoadBalancer":
		svcType = corev1.ServiceTypeLoadBalancer
	case "ExternalName":
		svcType = corev1.ServiceTypeExternalName
	}

	var ports []corev1.ServicePort
	for _, p := range cfg.Ports {
		proto := corev1.ProtocolTCP
		if p.Protocol == "UDP" {
			proto = corev1.ProtocolUDP
		}
		ports = append(ports, corev1.ServicePort{
			Name:       p.Name,
			Port:       int32(p.Port),
			TargetPort: intstr.FromInt32(int32(p.TargetPort)),
			Protocol:   proto,
		})
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:     svcType,
			Selector: cfg.Selector,
			Ports:    ports,
		},
	}

	out, err := yaml.Marshal(svc)
	if err != nil {
		return "", fmt.Errorf("序列化 Service YAML 失败: %w", err)
	}
	return string(out), nil
}

// buildIngressYAML 构建 K8s Ingress YAML
func buildIngressYAML(name, namespace string, config datatypes.JSON) (string, error) {
	var cfg IngressConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return "", fmt.Errorf("解析 Ingress 配置失败: %w", err)
	}

	var tls []networkingv1.IngressTLS
	for _, t := range cfg.TLS {
		tls = append(tls, networkingv1.IngressTLS{
			Hosts:      t.Hosts,
			SecretName: t.SecretName,
		})
	}

	var rules []networkingv1.IngressRule
	for _, r := range cfg.Rules {
		var paths []networkingv1.HTTPIngressPath
		for _, p := range r.Paths {
			pt := networkingv1.PathTypePrefix
			switch p.PathType {
			case "Exact":
				pt = networkingv1.PathTypeExact
			case "ImplementationSpecific":
				pt = networkingv1.PathTypeImplementationSpecific
			}
			paths = append(paths, networkingv1.HTTPIngressPath{
				Path:     p.Path,
				PathType: &pt,
				Backend: networkingv1.IngressBackend{
					Service: &networkingv1.IngressServiceBackend{
						Name: p.Backend.ServiceName,
						Port: networkingv1.ServiceBackendPort{Number: int32(p.Backend.ServicePort)},
					},
				},
			})
		}
		rules = append(rules, networkingv1.IngressRule{
			Host: r.Host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{Paths: paths},
			},
		})
	}

	ingress := &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{APIVersion: "networking.k8s.io/v1", Kind: "Ingress"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: cfg.Annotations,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &cfg.IngressClassName,
			TLS:              tls,
			Rules:            rules,
		},
	}

	out, err := yaml.Marshal(ingress)
	if err != nil {
		return "", fmt.Errorf("序列化 Ingress YAML 失败: %w", err)
	}
	return string(out), nil
}

// buildIngressRouteYAML 构建 Traefik IngressRoute YAML（CRD，使用 Unstructured）
func buildIngressRouteYAML(name, namespace string, config datatypes.JSON) (string, error) {
	var cfg IngressRouteConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return "", fmt.Errorf("解析 IngressRoute 配置失败: %w", err)
	}

	var routes []interface{}
	for _, r := range cfg.Routes {
		// 后端服务列表
		var services []interface{}
		for _, s := range r.Services {
			svcMap := map[string]interface{}{
				"name": s.Name,
				"port": int64(s.Port),
			}
			if s.Namespace != "" {
				svcMap["namespace"] = s.Namespace
			}
			if s.Kind != "" {
				svcMap["kind"] = s.Kind
			}
			if s.PassHostHeader != nil && !*s.PassHostHeader {
				svcMap["passHostHeader"] = false
			}
			if s.Scheme != "" {
				svcMap["scheme"] = s.Scheme
			}
			if s.NativeLB {
				svcMap["nativeLB"] = true
			}
			if s.Weight > 0 {
				svcMap["weight"] = int64(s.Weight)
			}
			services = append(services, svcMap)
		}

		// 路由条目
		routeKind := r.Kind
		if routeKind == "" {
			routeKind = "Rule"
		}
		entry := map[string]interface{}{
			"kind":     routeKind,
			"match":    r.Match,
			"services": services,
		}
		if r.Priority > 0 {
			entry["priority"] = int64(r.Priority)
		}

		// 中间件引用
		if len(r.Middlewares) > 0 {
			var mws []interface{}
			for _, mw := range r.Middlewares {
				mwMap := map[string]interface{}{"name": mw.Name}
				if mw.Namespace != "" {
					mwMap["namespace"] = mw.Namespace
				}
				mws = append(mws, mwMap)
			}
			entry["middlewares"] = mws
		}

		routes = append(routes, entry)
	}

	spec := map[string]interface{}{
		"routes": routes,
	}
	if len(cfg.EntryPoints) > 0 {
		spec["entryPoints"] = toInterfaceSlice(cfg.EntryPoints)
	}

	// TLS 配置
	if cfg.TLS != nil {
		tlsMap := map[string]interface{}{}
		if cfg.TLS.SecretName != "" {
			tlsMap["secretName"] = cfg.TLS.SecretName
		}
		if cfg.TLS.CertResolver != "" {
			tlsMap["certResolver"] = cfg.TLS.CertResolver
		}
		if len(cfg.TLS.Domains) > 0 {
			var domains []interface{}
			for _, d := range cfg.TLS.Domains {
				dm := map[string]interface{}{"main": d.Main}
				if len(d.SANs) > 0 {
					dm["sans"] = toInterfaceSlice(d.SANs)
				}
				domains = append(domains, dm)
			}
			tlsMap["domains"] = domains
		}
		if cfg.TLS.Options != nil {
			optMap := map[string]interface{}{"name": cfg.TLS.Options.Name}
			if cfg.TLS.Options.Namespace != "" {
				optMap["namespace"] = cfg.TLS.Options.Namespace
			}
			tlsMap["options"] = optMap
		}
		if len(tlsMap) > 0 {
			spec["tls"] = tlsMap
		}
	}

	metadata := map[string]interface{}{
		"name":      name,
		"namespace": namespace,
	}
	if len(cfg.Annotations) > 0 {
		metadata["annotations"] = toStringInterfaceMap(cfg.Annotations)
	}

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "traefik.io/v1alpha1",
			"kind":       "IngressRoute",
			"metadata":   metadata,
			"spec":       spec,
		},
	}

	out, err := yaml.Marshal(obj.Object)
	if err != nil {
		return "", fmt.Errorf("序列化 IngressRoute YAML 失败: %w", err)
	}
	return string(out), nil
}

// buildApisixRouteYAML 构建 APISIX Route YAML（CRD，使用 Unstructured）
func buildApisixRouteYAML(name, namespace string, config datatypes.JSON) (string, error) {
	var cfg ApisixRouteConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return "", fmt.Errorf("解析 ApisixRoute 配置失败: %w", err)
	}

	var httpRules []interface{}
	for _, rule := range cfg.HTTP {
		// match 条件
		matchMap := map[string]interface{}{
			"paths": toInterfaceSlice(rule.Match.Paths),
		}
		if len(rule.Match.Hosts) > 0 {
			matchMap["hosts"] = toInterfaceSlice(rule.Match.Hosts)
		}
		if len(rule.Match.Methods) > 0 {
			matchMap["methods"] = toInterfaceSlice(rule.Match.Methods)
		}
		if len(rule.Match.RemoteAddrs) > 0 {
			matchMap["remoteAddrs"] = toInterfaceSlice(rule.Match.RemoteAddrs)
		}

		// backends
		var backends []interface{}
		for _, b := range rule.Backends {
			bMap := map[string]interface{}{
				"serviceName": b.ServiceName,
				"servicePort": int64(b.ServicePort),
			}
			if b.Weight > 0 {
				bMap["weight"] = int64(b.Weight)
			}
			if b.ResolveGranularity != "" {
				bMap["resolveGranularity"] = b.ResolveGranularity
			}
			if b.Subset != "" {
				bMap["subset"] = b.Subset
			}
			backends = append(backends, bMap)
		}

		httpRule := map[string]interface{}{
			"name":     rule.Name,
			"match":    matchMap,
			"backends": backends,
		}
		if rule.Priority > 0 {
			httpRule["priority"] = int64(rule.Priority)
		}

		// plugins
		if len(rule.Plugins) > 0 {
			var plugins []interface{}
			for _, pl := range rule.Plugins {
				pluginEntry := map[string]interface{}{
					"name":   pl.Name,
					"enable": pl.Enable,
				}
				if pl.Config != nil {
					pluginEntry["config"] = pl.Config
				}
				plugins = append(plugins, pluginEntry)
			}
			httpRule["plugins"] = plugins
		}

		// timeout
		if rule.Timeout != nil {
			tm := map[string]interface{}{}
			if rule.Timeout.Connect != "" {
				tm["connect"] = rule.Timeout.Connect
			}
			if rule.Timeout.Read != "" {
				tm["read"] = rule.Timeout.Read
			}
			if rule.Timeout.Send != "" {
				tm["send"] = rule.Timeout.Send
			}
			if len(tm) > 0 {
				httpRule["timeout"] = tm
			}
		}

		if rule.PluginConfigName != "" {
			httpRule["plugin_config_name"] = rule.PluginConfigName
		}

		httpRules = append(httpRules, httpRule)
	}

	metadata := map[string]interface{}{
		"name":      name,
		"namespace": namespace,
	}
	if len(cfg.Annotations) > 0 {
		metadata["annotations"] = toStringInterfaceMap(cfg.Annotations)
	}

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apisix.apache.org/v2",
			"kind":       "ApisixRoute",
			"metadata":   metadata,
			"spec": map[string]interface{}{
				"http": httpRules,
			},
		},
	}

	out, err := yaml.Marshal(obj.Object)
	if err != nil {
		return "", fmt.Errorf("序列化 ApisixRoute YAML 失败: %w", err)
	}
	return string(out), nil
}

// buildApisixUpstreamYAML 构建 APISIX Upstream YAML（CRD，用于 gRPC 等场景）
func buildApisixUpstreamYAML(name, namespace string, config datatypes.JSON) (string, error) {
	var cfg ApisixUpstreamConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return "", fmt.Errorf("解析 ApisixUpstream 配置失败: %w", err)
	}

	spec := map[string]interface{}{}

	if cfg.Scheme != "" {
		spec["scheme"] = cfg.Scheme
	}
	if cfg.LoadBalancer != nil {
		lb := map[string]interface{}{"type": cfg.LoadBalancer.Type}
		if cfg.LoadBalancer.HashOn != "" {
			lb["hashOn"] = cfg.LoadBalancer.HashOn
		}
		if cfg.LoadBalancer.Key != "" {
			lb["key"] = cfg.LoadBalancer.Key
		}
		spec["loadbalancer"] = lb
	}
	if cfg.Retries > 0 {
		spec["retries"] = int64(cfg.Retries)
	}
	if cfg.Timeout != nil {
		tm := map[string]interface{}{}
		if cfg.Timeout.Connect != "" {
			tm["connect"] = cfg.Timeout.Connect
		}
		if cfg.Timeout.Read != "" {
			tm["read"] = cfg.Timeout.Read
		}
		if cfg.Timeout.Send != "" {
			tm["send"] = cfg.Timeout.Send
		}
		if len(tm) > 0 {
			spec["timeout"] = tm
		}
	}
	if cfg.HealthCheck != nil {
		spec["healthCheck"] = cfg.HealthCheck
	}
	if len(cfg.PortLevelSettings) > 0 {
		var pls []interface{}
		for _, p := range cfg.PortLevelSettings {
			plMap := map[string]interface{}{
				"port":   int64(p.Port),
				"scheme": p.Scheme,
			}
			pls = append(pls, plMap)
		}
		spec["portLevelSettings"] = pls
	}

	metadata := map[string]interface{}{
		"name":      name,
		"namespace": namespace,
	}
	if len(cfg.Annotations) > 0 {
		metadata["annotations"] = toStringInterfaceMap(cfg.Annotations)
	}

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apisix.apache.org/v2",
			"kind":       "ApisixUpstream",
			"metadata":   metadata,
			"spec":       spec,
		},
	}

	out, err := yaml.Marshal(obj.Object)
	if err != nil {
		return "", fmt.Errorf("序列化 ApisixUpstream YAML 失败: %w", err)
	}
	return string(out), nil
}

// --- 辅助函数 ---

func toInterfaceSlice(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

func toStringInterfaceMap(m map[string]string) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
