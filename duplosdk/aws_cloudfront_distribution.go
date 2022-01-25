package duplosdk

import (
	"fmt"
	"time"
)

type DuploAwsCloudfrontDefaultCacheBehavior struct {
	AllowedMethods             *DuploCFDAllowedMethods                       `json:"AllowedMethods,omitempty"`
	CachePolicyId              string                                        `json:"CachePolicyId,omitempty"`
	Compress                   bool                                          `json:"Compress"`
	DefaultTTL                 int                                           `json:"DefaultTTL,omitempty"`
	FieldLevelEncryptionId     string                                        `json:"FieldLevelEncryptionId"`
	OriginRequestPolicyId      string                                        `json:"OriginRequestPolicyId,omitempty"`
	LambdaFunctionAssociations *DuploAwsCloudfrontLambdaFunctionAssociations `json:"LambdaFunctionAssociations"`
	MaxTTL                     int                                           `json:"MaxTTL,omitempty"`
	MinTTL                     int                                           `json:"MinTTL,omitempty"`
	SmoothStreaming            bool                                          `json:"SmoothStreaming"`
	TargetOriginId             string                                        `json:"TargetOriginId,omitempty"`
	TrustedSigners             *DuploCFDTrustedSigners                       `json:"TrustedSigners,omitempty"`
	ViewerProtocolPolicy       *DuploStringValue                             `json:"ViewerProtocolPolicy,omitempty"`
	ForwardedValues            *DuploCFDForwardedValues                      `json:"ForwardedValues,omitempty"`
}

type DuploAwsCloudfrontCacheBehavior struct {
	AllowedMethods             *DuploCFDAllowedMethods                       `json:"AllowedMethods,omitempty"`
	CachePolicyId              string                                        `json:"CachePolicyId,omitempty"`
	Compress                   bool                                          `json:"Compress"`
	DefaultTTL                 int                                           `json:"DefaultTTL,omitempty"`
	FieldLevelEncryptionId     string                                        `json:"FieldLevelEncryptionId"`
	OriginRequestPolicyId      string                                        `json:"OriginRequestPolicyId,omitempty"`
	LambdaFunctionAssociations *DuploAwsCloudfrontLambdaFunctionAssociations `json:"LambdaFunctionAssociations"`
	MaxTTL                     int                                           `json:"MaxTTL,omitempty"`
	MinTTL                     int                                           `json:"MinTTL,omitempty"`
	SmoothStreaming            bool                                          `json:"SmoothStreaming"`
	TargetOriginId             string                                        `json:"TargetOriginId"`
	TrustedSigners             *DuploCFDTrustedSigners                       `json:"TrustedSigners,omitempty"`
	ViewerProtocolPolicy       *DuploStringValue                             `json:"ViewerProtocolPolicy,omitempty"`
	ForwardedValues            *DuploCFDForwardedValues                      `json:"ForwardedValues,omitempty"`
	PathPattern                string                                        `json:"PathPattern"`
}

type DuploAwsCloudfrontLambdaFunctionAssociations struct {
	Items    *[]DuploAwsCloudfrontLambdaFunctionAssociation `json:"Items"`
	Quantity int                                            `json:"Quantity"`
}

type DuploAwsCloudfrontLambdaFunctionAssociation struct {
	EventType         string `json:"EventType"`
	LambdaFunctionARN string `json:"LambdaFunctionARN"`
	IncludeBody       bool   `json:"IncludeBody"`
}

type DuploAwsCloudfrontCacheBehaviors struct {
	Items    *[]DuploAwsCloudfrontCacheBehavior `json:"Items"`
	Quantity int                                `json:"Quantity"`
}
type DuploCFDAllowedMethods struct {
	CachedMethods *DuploCFDStringItems `json:"CachedMethods,omitempty"`
	Items         []string             `json:"Items"`
	Quantity      int                  `json:"Quantity"`
}

type DuploCFDCookiePreference struct {
	Forward          DuploStringValue     `json:"Forward,omitempty"`
	WhitelistedNames *DuploCFDStringItems `json:"WhitelistedNames,omitempty"`
}

type DuploCFDStringItems struct {
	Items    []string `json:"Items"`
	Quantity int      `json:"Quantity"`
}

type DuploCFDForwardedValues struct {
	Cookies              *DuploCFDCookiePreference `json:"Cookies,omitempty"`
	Headers              *DuploCFDStringItems      `json:"Headers,omitempty"`
	QueryString          bool                      `json:"QueryString"`
	QueryStringCacheKeys *DuploCFDStringItems      `json:"QueryStringCacheKeys,omitempty"`
}

type DuploCFDTrustedSigners struct {
	Enabled  bool     `json:"Enabled"`
	Items    []string `json:"Items"`
	Quantity int      `json:"Quantity"`
}

type DuploAwsCloudfrontOrigins struct {
	Items    *[]DuploAwsCloudfrontOrigin `json:"Items"`
	Quantity int                         `json:"Quantity"`
}

type DuploAwsCloudfrontOrigin struct {
	ConnectionAttempts int                                    `json:"ConnectionAttempts,omitempty"`
	ConnectionTimeout  int                                    `json:"ConnectionTimeout,omitempty"`
	CustomHeaders      *DuploAwsCloudfrontOriginCustomHeaders `json:"CustomHeaders,omitempty"`
	DomainName         string                                 `json:"DomainName,omitempty"`
	Id                 string                                 `json:"Id,omitempty"`
	OriginPath         string                                 `json:"OriginPath"`
	S3OriginConfig     *DuploAwsCloudfrontOriginS3Config      `json:"S3OriginConfig,omitempty"`
	CustomOriginConfig *DuploAwsCloudfrontCustomOriginConfig  `json:"CustomOriginConfig,omitempty"`
}

type DuploAwsCloudfrontOriginCustomHeaders struct {
	Items    *[]DuploAwsCloudfrontOriginCustomHeader `json:"Items"`
	Quantity int                                     `json:"Quantity"`
}

type DuploAwsCloudfrontOriginCustomHeader struct {
	HeaderName  string `json:"HeaderName,omitempty"`
	HeaderValue string `json:"HeaderValue,omitempty"`
}

type DuploAwsCloudfrontOriginS3Config struct {
	OriginAccessIdentity string `json:"OriginAccessIdentity"`
}

type DuploAwsCloudfrontCustomOriginConfig struct {
	HTTPPort               int                  `json:"HTTPPort,omitempty"`
	HTTPSPort              int                  `json:"HTTPSPort,omitempty"`
	OriginKeepaliveTimeout int                  `json:"OriginKeepaliveTimeout,omitempty"`
	OriginReadTimeout      int                  `json:"OriginReadTimeout,omitempty"`
	OriginProtocolPolicy   *DuploStringValue    `json:"OriginProtocolPolicy,omitempty"`
	OriginSslProtocols     *DuploCFDStringItems `json:"OriginSslProtocols,omitempty"`
}

type DuploAwsCloudfrontDistributionRestrictions struct {
	GeoRestriction *DuploAwsCloudfrontDistributionGeoRestriction `json:"GeoRestriction,omitempty"`
}

type DuploAwsCloudfrontDistributionGeoRestriction struct {
	RestrictionType *DuploStringValue `json:"RestrictionType,omitempty"`
	Items           []string          `json:"Items"`
	Quantity        int               `json:"Quantity"`
}

type DuploAwsCloudfrontDistributionViewerCertificate struct {
	ACMCertificateArn            string            `json:"ACMCertificateArn,omitempty"`
	CloudFrontDefaultCertificate bool              `json:"CloudFrontDefaultCertificate,omitempty"`
	MinimumProtocolVersion       *DuploStringValue `json:"MinimumProtocolVersion,omitempty"`
	SSLSupportMethod             *DuploStringValue `json:"SSLSupportMethod,omitempty"`
	IAMCertificateId             string            `json:"IAMCertificateId,omitempty"`
}
type DuploAwsCloudfrontDistributionOriginGroups struct {
	Items    *[]DuploAwsCloudfrontDistributionOriginGroup `json:"Items"`
	Quantity int                                          `json:"Quantity"`
}

type DuploAwsCloudfrontDistributionOriginGroup struct {
	FailoverCriteria *DuploOriginGroupFailoverCriteria                 `json:"FailoverCriteria,omitempty"`
	Id               string                                            `json:"Id,omitempty"`
	Members          *DuploAwsCloudfrontDistributionOriginGroupMembers `json:"Members,omitempty"`
}

type DuploOriginGroupFailoverCriteriaStatusCodes struct {
	Items    []int `json:"Items"`
	Quantity int   `json:"Quantity"`
}

type DuploOriginGroupFailoverCriteria struct {
	StatusCodes *DuploOriginGroupFailoverCriteriaStatusCodes `json:"StatusCodes,omitempty"`
}

type DuploAwsCloudfrontDistributionOriginGroupMembers struct {
	Items    *[]DuploAwsCloudfrontDistributionOriginGroupMember `json:"Items"`
	Quantity int                                                `json:"Quantity"`
}

type DuploAwsCloudfrontDistributionOriginGroupMember struct {
	OriginId string `json:"OriginId,omitempty"`
}

type DuploAwsCloudfrontDistributionLoggingConfig struct {
	Bucket         string `json:"Bucket"`
	Enabled        bool   `json:"Enabled"`
	IncludeCookies bool   `json:"IncludeCookies"`
	Prefix         string `json:"Prefix"`
}

type DuploAwsCloudfrontDistributionCustomErrorResponses struct {
	Items    *[]DuploAwsCloudfrontDistributionCustomErrorResponse `json:"Items"`
	Quantity int                                                  `json:"Quantity,"`
}

type DuploAwsCloudfrontDistributionCustomErrorResponse struct {
	ErrorCachingMinTTL int    `json:"ErrorCachingMinTTL,omitempty"`
	ErrorCode          int    `json:"ErrorCode,omitempty"`
	ResponseCode       string `json:"ResponseCode"`
	ResponsePagePath   string `json:"ResponsePagePath,omitempty"`
}

type DuploAwsCloudfrontDistributionCreate struct {
	DistributionConfig *DuploAwsCloudfrontDistributionConfig `json:"DistributionConfig,omitempty"`
	Id                 string                                `json:"Id,omitempty"`
	IfMatch            string                                `json:"IfMatch,omitempty"`
	UseOAIIdentity     bool                                  `json:"UseOAIIdentity,omitempty"`
}

type DuploAwsCloudfrontDistribution struct {
	DistributionConfig *DuploAwsCloudfrontDistributionConfig `json:"DistributionConfig,omitempty"`
	Id                 string                                `json:"Id,omitempty"`
	ARN                string                                `json:"ARN,omitempty"`
	DomainName         string                                `json:"DomainName,omitempty"`
	Status             string                                `json:"Status,omitempty"`
}

type DuploAwsCloudfrontDistributionGetResponse struct {
	Distribution *DuploAwsCloudfrontDistribution `json:"Distribution,omitempty"`
	ETag         string                          `json:"ETag,omitempty"`
}

type DuploAwsCloudfrontDistributionDisable struct {
	Id                 string                               `json:"Id,omitempty"`
	DistributionConfig DuploAwsCloudfrontDistributionConfig `json:"DistributionConfig,omitempty"`
}

type DuploAwsCloudfrontDistributionConfig struct {
	Aliases           *DuploCFDStringItems `json:"Aliases,omitempty"`
	AliasICPRecordals []struct {
		CNAME             string `json:"CNAME,omitempty"`
		ICPRecordalStatus struct {
			Value string `json:"Value,omitempty"`
		} `json:"ICPRecordalStatus,omitempty"`
	} `json:"AliasICPRecordals,omitempty"`
	ARN                  string                                              `json:"ARN,omitempty"`
	CacheBehaviors       *DuploAwsCloudfrontCacheBehaviors                   `json:"CacheBehaviors,omitempty"`
	Comment              string                                              `json:"Comment"`
	DefaultRootObject    string                                              `json:"DefaultRootObject"`
	CustomErrorResponses *DuploAwsCloudfrontDistributionCustomErrorResponses `json:"CustomErrorResponses,omitempty"`
	DefaultCacheBehavior *DuploAwsCloudfrontDefaultCacheBehavior             `json:"DefaultCacheBehavior,omitempty"`
	DomainName           string                                              `json:"DomainName,omitempty"`
	Enabled              bool                                                `json:"Enabled"`
	HttpVersion          *DuploStringValue                                   `json:"HttpVersion,omitempty"`
	Id                   string                                              `json:"Id,omitempty"`
	IsIPV6Enabled        bool                                                `json:"IsIPV6Enabled"`
	LastModifiedTime     time.Time                                           `json:"LastModifiedTime,omitempty"`
	OriginGroups         *DuploAwsCloudfrontDistributionOriginGroups         `json:"OriginGroups,omitempty"`
	Origins              *DuploAwsCloudfrontOrigins                          `json:"Origins,omitempty"`
	PriceClass           *DuploStringValue                                   `json:"PriceClass,omitempty"`
	Restrictions         *DuploAwsCloudfrontDistributionRestrictions         `json:"Restrictions,omitempty"`
	Status               string                                              `json:"Status,omitempty"`
	ViewerCertificate    *DuploAwsCloudfrontDistributionViewerCertificate    `json:"ViewerCertificate,omitempty"`
	Logging              *DuploAwsCloudfrontDistributionLoggingConfig        `json:"Logging,omitempty"`
	WebACLId             string                                              `json:"WebACLId"`
}

func (c *Client) AwsCloudfrontDistributionCreate(tenantID string, rq *DuploAwsCloudfrontDistributionCreate) (*DuploAwsCloudfrontDistribution, ClientError) {
	resp := DuploAwsCloudfrontDistribution{}
	err := c.postAPI(
		fmt.Sprintf("AwsCloudfrontDistributionCreate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/cloudFrontDistribution", tenantID),
		&rq,
		&resp,
	)
	return &resp, err
}

func (c *Client) AwsCloudfrontDistributionUpdate(tenantID string, rq *DuploAwsCloudfrontDistributionCreate) (*DuploAwsCloudfrontDistribution, ClientError) {
	resp := DuploAwsCloudfrontDistribution{}
	err := c.putAPI(
		fmt.Sprintf("AwsCloudfrontDistributionUpdate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/cloudFrontDistribution", tenantID),
		&rq,
		&resp,
	)
	return &resp, err
}

func (c *Client) AwsCloudfrontDistributionGet(tenantID string, cfdId string) (*DuploAwsCloudfrontDistributionGetResponse, ClientError) {
	rp := DuploAwsCloudfrontDistributionGetResponse{}
	err := c.getAPI(
		fmt.Sprintf("AwsCloudfrontDistributionGet(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/cloudFrontDistribution/%s", tenantID, cfdId),
		&rp,
	)
	return &rp, err
}

func (c *Client) AwsCloudfrontDistributionGetFromList(tenantID string, cfdId string) (*DuploAwsCloudfrontDistributionConfig, ClientError) {
	list, err := c.AwsCloudfrontDistributionList(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, element := range *list {
			if element.Id == cfdId {
				return &element, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) AwsCloudfrontDistributionList(tenantID string) (*[]DuploAwsCloudfrontDistributionConfig, ClientError) {
	rp := []DuploAwsCloudfrontDistributionConfig{}
	err := c.getAPI(
		fmt.Sprintf("AwsCloudfrontDistributionList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/cloudFrontDistribution", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) AwsCloudfrontDistributionExists(tenantID, id string) (bool, ClientError) {
	list, err := c.AwsCloudfrontDistributionList(tenantID)
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, element := range *list {
			if element.Id == id {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) AwsCloudfrontDistributionDelete(tenantID string, id string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("AwsCloudfrontDistributionDelete(%s, %s)", tenantID, id),
		fmt.Sprintf("v3/subscriptions/%s/aws/cloudFrontDistribution/%s", tenantID, id),
		nil,
	)
}

func (c *Client) AwsCloudfrontDistributionDisable(tenantID string, id string) ClientError {
	rp := DuploAwsCloudfrontDistribution{}
	rq := map[string]interface{}{
		"Id": id,
		"DistributionConfig": map[string]bool{
			"Enabled": false,
		},
	}
	return c.putAPI(
		fmt.Sprintf("AwsCloudfrontDistributionDisable(%s, %s)", tenantID, id),
		fmt.Sprintf("v3/subscriptions/%s/aws/cloudFrontDistribution/%s", tenantID, id),
		&rq,
		&rp,
	)
}
