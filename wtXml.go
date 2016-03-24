////////////////////////////////////////////////////////////////////////////
// Porgram: wtXml - web test XML file structs
// authors: Antonio Sun (c) 2015-16, All rights reserved
////////////////////////////////////////////////////////////////////////////

package main

////////////////////////////////////////////////////////////////////////////
// Constant and data type/structure definitions

type Xml struct {
	Xml string `xml:",innerxml"`
}

type XmlBase struct {
	Name           string `xml:"DisplayName,attr"`
	RuleParameters Xml
}

/*
   <Comment CommentText="[#30]" />
*/
type Comment struct {
	Comment string `xml:"CommentText,attr"`
}

/*
  <ContextParameters>
    <ContextParameter Name="web" Value="http://localhost:50357/" />
  </ContextParameters>
*/
type ContextParameter struct {
	Name  string `xml:"Name,attr"`
	Value string `xml:"Value,attr"`
}

/*
   <DataSource Name="DataSource1" Provider="Microsoft.VisualStudio.TestTools.DataSource.CSV" Connection="|DataDirectory|\.\Data\text.csv">
     <Tables>
       <DataSourceTable Name="text#csv" SelectColumns="SelectOnlyBoundColumns" AccessMethod="Sequential" />
     </Tables>
   </DataSource>
*/
type DataSource struct {
	Name       string `xml:"Name,attr"`
	Connection string `xml:"Connection,attr"`
	Tables     struct {
		DataSourceTable string `xml:",innerxml"`
	}
}

/*
   <ConditionalRule Classname="Microsoft.VisualStudio.TestTools.WebTesting.Rules.NumericalComparisonRule, Microsoft.VisualStudio.QualityTools.WebTestFramework, Version=10.0.0.0, Culture=neutral, PublicKeyToken=b03f5f7f11d50a3a" DisplayName="Number Comparison" Description="The condition is met when the value of the context parameter satisfies the comparison with the provided value.">
     <RuleParameters>
       <RuleParameter Name="ContextParameterName" Value="Ver" />
       <RuleParameter Name="ComparisonOperator" Value="==" />
       <RuleParameter Name="Value" Value="2.1" />
     </RuleParameters>
   </ConditionalRule>
*/
type ConditionalRule struct {
	XmlBase
}

// <IncludedWebTest Name="..." ...
type IncludedWebTest struct {
	Included string `xml:"Name,attr"`
}

/*
  <ValidationRules>
    <ValidationRule Classname="Microsoft.VisualStudio.TestTools.WebTesting.Rules.ValidateResponseUrl, Microsoft.VisualStudio.QualityTools.WebTestFramework, Version=10.0.0.0, Culture=neutral, PublicKeyToken=b03f5f7f11d50a3a" DisplayName="Response URL" Description="Validates that the response URL after redirects are followed is the same as the recorded response URL.  QueryString parameters are ignored." Level="Low" ExectuionOrder="BeforeDependents" />
    <ValidationRule Classname="Microsoft.VisualStudio.TestTools.WebTesting.Rules.ValidationRuleResponseTimeGoal, Microsoft.VisualStudio.QualityTools.WebTestFramework, Version=10.0.0.0, Culture=neutral, PublicKeyToken=b03f5f7f11d50a3a" DisplayName="Response Time Goal" Description="Validates that the response time for the request is less than or equal to the response time goal as specified on the request.  Response time goals of zero will be ignored." Level="Low" ExectuionOrder="AfterDependents">
      <RuleParameters>
        <RuleParameter Name="Tolerance" Value="0" />
      </RuleParameters>
    </ValidationRule>
  </ValidationRules>
*/
type ValidationRules struct {
	ValidationRule []struct {
		XmlBase
	}
}

type Request struct {
	/*
	   <Request Method="GET" Version="1.1" Url="{{web}}Account/LogOn" ThinkTime="0" Timeout="300" ParseDependentRequests="True" FollowRedirects="True" RecordResult="True" Cache="False" ResponseTimeGoal="0" Encoding="utf-8" ExpectedHttpStatusCode="0" ExpectedResponseUrl="" ReportingName="">
	*/
	Url           string `xml:"Url,attr"`
	ThinkTime     string `xml:"ThinkTime,attr"`
	Timeout       string `xml:"Timeout,attr"`
	RecordResult  string `xml:"RecordResult,attr"`
	ReportingName string `xml:"ReportingName,attr"`

	/*
	   <RequestPlugins>
	     <RequestPlugin Classname="Microsoft.VisualStudio.WebTesting.PowerTools.SharePoint.MTSL.General.SPLTPT_MTSL_SetContextParameterValue, Microsoft.VisualStudio.QualityTools.WebTestFramework, Version=10.0.0.0, Culture=neutral, PublicKeyToken=b03f5f7f11d50a3a" DisplayName="Set Context Parameter Value" Description="Allows you to set a context parameter value for this request.">
	       <RuleParameters>
	         <RuleParameter Name="Enabled" Value="True" />
	         <RuleParameter Name="sContextParameterName" Value="deviceName" />
	         <RuleParameter Name="sContextParameterValue" Value="DevA" />
	         <RuleParameter Name="bDoReplace" Value="False" />
	         <RuleParameter Name="sReplaceFindPattern" Value="" />
	         <RuleParameter Name="sReplaceWith" Value="" />
	         <RuleParameter Name="bUseRegEx" Value="False" />
	         <RuleParameter Name="bApplyBeforeRequest" Value="True" />
	         <RuleParameter Name="bHTMLEncode" Value="False" />
	         <RuleParameter Name="bHTMLDecode" Value="False" />
	         <RuleParameter Name="bURLEncode" Value="False" />
	         <RuleParameter Name="bURLDecode" Value="False" />
	         <RuleParameter Name="bBase64Encode" Value="False" />
	         <RuleParameter Name="bBase64Decode" Value="False" />
	         <RuleParameter Name="bRemoveUnicodeEscapeSequences" Value="False" />
	       </RuleParameters>
	     </RequestPlugin>
	     <RequestPlugin Classname="Microsoft.VisualStudio.WebTesting.PowerTools.SharePoint.MTSL.General.SPLTPT_MTSL_SetContextParameterValue, Microsoft.VisualStudio.QualityTools.WebTestFramework, Version=10.0.0.0, Culture=neutral, PublicKeyToken=b03f5f7f11d50a3a" DisplayName="Set Context Parameter Value" Description="Allows you to set a context parameter value for this request.">
	       <RuleParameters>
	         <RuleParameter Name="Enabled" Value="True" />
	         <RuleParameter Name="sContextParameterName" Value="deviceName2" />
	         <RuleParameter Name="sContextParameterValue" Value="DevB" />
	         <RuleParameter Name="bDoReplace" Value="False" />
	         <RuleParameter Name="sReplaceFindPattern" Value="" />
	         <RuleParameter Name="sReplaceWith" Value="" />
	         <RuleParameter Name="bUseRegEx" Value="False" />
	         <RuleParameter Name="bApplyBeforeRequest" Value="True" />
	         <RuleParameter Name="bHTMLEncode" Value="False" />
	         <RuleParameter Name="bHTMLDecode" Value="False" />
	         <RuleParameter Name="bURLEncode" Value="False" />
	         <RuleParameter Name="bURLDecode" Value="False" />
	         <RuleParameter Name="bBase64Encode" Value="False" />
	         <RuleParameter Name="bBase64Decode" Value="False" />
	         <RuleParameter Name="bRemoveUnicodeEscapeSequences" Value="False" />
	       </RuleParameters>
	     </RequestPlugin>
	   </RequestPlugins>
	*/
	RequestPlugins struct {
		RequestPlugin []struct {
			XmlBase
		}
	}

	/*
	   <ExtractionRules>
	     <ExtractionRule Classname="Microsoft.VisualStudio.TestTools.WebTesting.Rules.ExtractHiddenFields, Microsoft.VisualStudio.QualityTools.WebTestFramework, Version=10.0.0.0, Culture=neutral, PublicKeyToken=b03f5f7f11d50a3a" VariableName="1" DisplayName="Extract Hidden Fields" Description="Extract all hidden fields from the response and place them into the test context.">
	       <RuleParameters>
	         <RuleParameter Name="Required" Value="True" />
	         <RuleParameter Name="HtmlDecode" Value="True" />
	       </RuleParameters>
	     </ExtractionRule>
	   </ExtractionRules>
	*/
	ExtractionRules struct {
		ExtractionRule []struct {
			XmlBase
			VariableName string `xml:"VariableName,attr"`
		}
	}

	ValidationRules ValidationRules

	/* The QueryStringParameters actually belongs to GetRequest
		 but put here for convenience handling in dealReqAddons

	   <QueryStringParameters>
	     <QueryStringParameter Name="v" Value="SoftwareVersion" RecordedValue="" CorrelationBinding="" UrlEncode="True" UseToGroupResults="False" />
	     <QueryStringParameter Name="xref" Value="RRR" RecordedValue="" CorrelationBinding="" UrlEncode="True" UseToGroupResults="False" />
	   </QueryStringParameters>
	*/
	QueryStringParameters struct {
		QueryStringParameter string `xml:",innerxml"`
	}

	/* The QueryStringParameters actually belongs to PostRequest
		 but put here as well.

	   <FormPostHttpBody>
	     <FormPostParameter Name="N" Value="A=" RecordedValue="" CorrelationBinding="" UrlEncode="False" />
	   </FormPostHttpBody>
	*/
	FormPostHttpBody struct {
		FormPostParameter string `xml:",innerxml"`
	}
}

type GetRequest struct {
	Request
}

type PostRequest struct {
	Request
	StringBody string `xml:"StringHttpBody"`
}
