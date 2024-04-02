package constants

// DnsFormatAnnotation is the annotation used to determine the DNS format
type DnsFormatAnnotation string

const (
	DnsFormatAnnotationApiGateway DnsFormatAnnotation = "bkch.service-discovery.api-gateway"
	DnsFormatAnnotationInCluster  DnsFormatAnnotation = "bkch.service-discovery.in-cluster"
)

// String returns the string representation of the DnsFormatAnnotation
func (d DnsFormatAnnotation) String() string {
	return string(d)
}
