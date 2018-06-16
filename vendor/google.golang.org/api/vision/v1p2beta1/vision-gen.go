// Package vision provides access to the Cloud Vision API.
//
// See https://cloud.google.com/vision/
//
// Usage example:
//
//   import "google.golang.org/api/vision/v1p2beta1"
//   ...
//   visionService, err := vision.New(oauthHttpClient)
package vision // import "google.golang.org/api/vision/v1p2beta1"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	context "golang.org/x/net/context"
	ctxhttp "golang.org/x/net/context/ctxhttp"
	gensupport "google.golang.org/api/gensupport"
	googleapi "google.golang.org/api/googleapi"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Always reference these packages, just in case the auto-generated code
// below doesn't.
var _ = bytes.NewBuffer
var _ = strconv.Itoa
var _ = fmt.Sprintf
var _ = json.NewDecoder
var _ = io.Copy
var _ = url.Parse
var _ = gensupport.MarshalJSON
var _ = googleapi.Version
var _ = errors.New
var _ = strings.Replace
var _ = context.Canceled
var _ = ctxhttp.Do

const apiId = "vision:v1p2beta1"
const apiName = "vision"
const apiVersion = "v1p2beta1"
const basePath = "https://vision.googleapis.com/"

// OAuth2 scopes used by this API.
const (
	// View and manage your data across Google Cloud Platform services
	CloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

	// Apply machine learning models to understand and label images
	CloudVisionScope = "https://www.googleapis.com/auth/cloud-vision"
)

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	s.Files = NewFilesService(s)
	s.Images = NewImagesService(s)
	return s, nil
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment

	Files *FilesService

	Images *ImagesService
}

func (s *Service) userAgent() string {
	if s.UserAgent == "" {
		return googleapi.UserAgent
	}
	return googleapi.UserAgent + " " + s.UserAgent
}

func NewFilesService(s *Service) *FilesService {
	rs := &FilesService{s: s}
	return rs
}

type FilesService struct {
	s *Service
}

func NewImagesService(s *Service) *ImagesService {
	rs := &ImagesService{s: s}
	return rs
}

type ImagesService struct {
	s *Service
}

// AnnotateFileResponse: Response to a single file annotation request. A
// file may contain one or more
// images, which individually have their own responses.
type AnnotateFileResponse struct {
	// InputConfig: Information about the file for which this response is
	// generated.
	InputConfig *InputConfig `json:"inputConfig,omitempty"`

	// Responses: Individual responses to images found within the file.
	Responses []*AnnotateImageResponse `json:"responses,omitempty"`

	// ForceSendFields is a list of field names (e.g. "InputConfig") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "InputConfig") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *AnnotateFileResponse) MarshalJSON() ([]byte, error) {
	type NoMethod AnnotateFileResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// AnnotateImageResponse: Response to an image annotation request.
type AnnotateImageResponse struct {
	// Context: If present, contextual information is needed to understand
	// where this image
	// comes from.
	Context *ImageAnnotationContext `json:"context,omitempty"`

	// CropHintsAnnotation: If present, crop hints have completed
	// successfully.
	CropHintsAnnotation *CropHintsAnnotation `json:"cropHintsAnnotation,omitempty"`

	// Error: If set, represents the error message for the operation.
	// Note that filled-in image annotations are guaranteed to be
	// correct, even when `error` is set.
	Error *Status `json:"error,omitempty"`

	// FaceAnnotations: If present, face detection has completed
	// successfully.
	FaceAnnotations []*FaceAnnotation `json:"faceAnnotations,omitempty"`

	// FullTextAnnotation: If present, text (OCR) detection or document
	// (OCR) text detection has
	// completed successfully.
	// This annotation provides the structural hierarchy for the OCR
	// detected
	// text.
	FullTextAnnotation *TextAnnotation `json:"fullTextAnnotation,omitempty"`

	// ImagePropertiesAnnotation: If present, image properties were
	// extracted successfully.
	ImagePropertiesAnnotation *ImageProperties `json:"imagePropertiesAnnotation,omitempty"`

	// LabelAnnotations: If present, label detection has completed
	// successfully.
	LabelAnnotations []*EntityAnnotation `json:"labelAnnotations,omitempty"`

	// LandmarkAnnotations: If present, landmark detection has completed
	// successfully.
	LandmarkAnnotations []*EntityAnnotation `json:"landmarkAnnotations,omitempty"`

	// LogoAnnotations: If present, logo detection has completed
	// successfully.
	LogoAnnotations []*EntityAnnotation `json:"logoAnnotations,omitempty"`

	// SafeSearchAnnotation: If present, safe-search annotation has
	// completed successfully.
	SafeSearchAnnotation *SafeSearchAnnotation `json:"safeSearchAnnotation,omitempty"`

	// TextAnnotations: If present, text (OCR) detection has completed
	// successfully.
	TextAnnotations []*EntityAnnotation `json:"textAnnotations,omitempty"`

	// WebDetection: If present, web detection has completed successfully.
	WebDetection *WebDetection `json:"webDetection,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Context") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Context") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *AnnotateImageResponse) MarshalJSON() ([]byte, error) {
	type NoMethod AnnotateImageResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// AsyncAnnotateFileResponse: The response for a single offline file
// annotation request.
type AsyncAnnotateFileResponse struct {
	// OutputConfig: The output location and metadata from
	// AsyncAnnotateFileRequest.
	OutputConfig *OutputConfig `json:"outputConfig,omitempty"`

	// ForceSendFields is a list of field names (e.g. "OutputConfig") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "OutputConfig") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *AsyncAnnotateFileResponse) MarshalJSON() ([]byte, error) {
	type NoMethod AsyncAnnotateFileResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// AsyncBatchAnnotateFilesResponse: Response to an async batch file
// annotation request.
type AsyncBatchAnnotateFilesResponse struct {
	// Responses: The list of file annotation responses, one for each
	// request in
	// AsyncBatchAnnotateFilesRequest.
	Responses []*AsyncAnnotateFileResponse `json:"responses,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Responses") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Responses") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *AsyncBatchAnnotateFilesResponse) MarshalJSON() ([]byte, error) {
	type NoMethod AsyncBatchAnnotateFilesResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Block: Logical element on the page.
type Block struct {
	// BlockType: Detected block type (text, image etc) for this block.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown block type.
	//   "TEXT" - Regular text block.
	//   "TABLE" - Table block.
	//   "PICTURE" - Image block.
	//   "RULER" - Horizontal/vertical line box.
	//   "BARCODE" - Barcode block.
	BlockType string `json:"blockType,omitempty"`

	// BoundingBox: The bounding box for the block.
	// The vertices are in the order of top-left, top-right,
	// bottom-right,
	// bottom-left. When a rotation of the bounding box is detected the
	// rotation
	// is represented as around the top-left corner as defined when the text
	// is
	// read in the 'natural' orientation.
	// For example:
	//
	// * when the text is horizontal it might look like:
	//
	//         0----1
	//         |    |
	//         3----2
	//
	// * when it's rotated 180 degrees around the top-left corner it
	// becomes:
	//
	//         2----3
	//         |    |
	//         1----0
	//
	//   and the vertice order will still be (0, 1, 2, 3).
	BoundingBox *BoundingPoly `json:"boundingBox,omitempty"`

	// Confidence: Confidence of the OCR results on the block. Range [0, 1].
	Confidence float64 `json:"confidence,omitempty"`

	// Paragraphs: List of paragraphs in this block (if this blocks is of
	// type text).
	Paragraphs []*Paragraph `json:"paragraphs,omitempty"`

	// Property: Additional information detected for the block.
	Property *TextProperty `json:"property,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BlockType") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BlockType") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Block) MarshalJSON() ([]byte, error) {
	type NoMethod Block
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *Block) UnmarshalJSON(data []byte) error {
	type NoMethod Block
	var s1 struct {
		Confidence gensupport.JSONFloat64 `json:"confidence"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	return nil
}

// BoundingPoly: A bounding polygon for the detected image annotation.
type BoundingPoly struct {
	// NormalizedVertices: The bounding polygon normalized vertices.
	NormalizedVertices []*NormalizedVertex `json:"normalizedVertices,omitempty"`

	// Vertices: The bounding polygon vertices.
	Vertices []*Vertex `json:"vertices,omitempty"`

	// ForceSendFields is a list of field names (e.g. "NormalizedVertices")
	// to unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "NormalizedVertices") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *BoundingPoly) MarshalJSON() ([]byte, error) {
	type NoMethod BoundingPoly
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Color: Represents a color in the RGBA color space. This
// representation is designed
// for simplicity of conversion to/from color representations in
// various
// languages over compactness; for example, the fields of this
// representation
// can be trivially provided to the constructor of "java.awt.Color" in
// Java; it
// can also be trivially provided to UIColor's
// "+colorWithRed:green:blue:alpha"
// method in iOS; and, with just a little work, it can be easily
// formatted into
// a CSS "rgba()" string in JavaScript, as well. Here are some
// examples:
//
// Example (Java):
//
//      import com.google.type.Color;
//
//      // ...
//      public static java.awt.Color fromProto(Color protocolor) {
//        float alpha = protocolor.hasAlpha()
//            ? protocolor.getAlpha().getValue()
//            : 1.0;
//
//        return new java.awt.Color(
//            protocolor.getRed(),
//            protocolor.getGreen(),
//            protocolor.getBlue(),
//            alpha);
//      }
//
//      public static Color toProto(java.awt.Color color) {
//        float red = (float) color.getRed();
//        float green = (float) color.getGreen();
//        float blue = (float) color.getBlue();
//        float denominator = 255.0;
//        Color.Builder resultBuilder =
//            Color
//                .newBuilder()
//                .setRed(red / denominator)
//                .setGreen(green / denominator)
//                .setBlue(blue / denominator);
//        int alpha = color.getAlpha();
//        if (alpha != 255) {
//          result.setAlpha(
//              FloatValue
//                  .newBuilder()
//                  .setValue(((float) alpha) / denominator)
//                  .build());
//        }
//        return resultBuilder.build();
//      }
//      // ...
//
// Example (iOS / Obj-C):
//
//      // ...
//      static UIColor* fromProto(Color* protocolor) {
//         float red = [protocolor red];
//         float green = [protocolor green];
//         float blue = [protocolor blue];
//         FloatValue* alpha_wrapper = [protocolor alpha];
//         float alpha = 1.0;
//         if (alpha_wrapper != nil) {
//           alpha = [alpha_wrapper value];
//         }
//         return [UIColor colorWithRed:red green:green blue:blue
// alpha:alpha];
//      }
//
//      static Color* toProto(UIColor* color) {
//          CGFloat red, green, blue, alpha;
//          if (![color getRed:&red green:&green blue:&blue
// alpha:&alpha]) {
//            return nil;
//          }
//          Color* result = [Color alloc] init];
//          [result setRed:red];
//          [result setGreen:green];
//          [result setBlue:blue];
//          if (alpha <= 0.9999) {
//            [result setAlpha:floatWrapperWithValue(alpha)];
//          }
//          [result autorelease];
//          return result;
//     }
//     // ...
//
//  Example (JavaScript):
//
//     // ...
//
//     var protoToCssColor = function(rgb_color) {
//        var redFrac = rgb_color.red || 0.0;
//        var greenFrac = rgb_color.green || 0.0;
//        var blueFrac = rgb_color.blue || 0.0;
//        var red = Math.floor(redFrac * 255);
//        var green = Math.floor(greenFrac * 255);
//        var blue = Math.floor(blueFrac * 255);
//
//        if (!('alpha' in rgb_color)) {
//           return rgbToCssColor_(red, green, blue);
//        }
//
//        var alphaFrac = rgb_color.alpha.value || 0.0;
//        var rgbParams = [red, green, blue].join(',');
//        return ['rgba(', rgbParams, ',', alphaFrac, ')'].join('');
//     };
//
//     var rgbToCssColor_ = function(red, green, blue) {
//       var rgbNumber = new Number((red << 16) | (green << 8) | blue);
//       var hexString = rgbNumber.toString(16);
//       var missingZeros = 6 - hexString.length;
//       var resultBuilder = ['#'];
//       for (var i = 0; i < missingZeros; i++) {
//          resultBuilder.push('0');
//       }
//       resultBuilder.push(hexString);
//       return resultBuilder.join('');
//     };
//
//     // ...
type Color struct {
	// Alpha: The fraction of this color that should be applied to the
	// pixel. That is,
	// the final pixel color is defined by the equation:
	//
	//   pixel color = alpha * (this color) + (1.0 - alpha) * (background
	// color)
	//
	// This means that a value of 1.0 corresponds to a solid color,
	// whereas
	// a value of 0.0 corresponds to a completely transparent color.
	// This
	// uses a wrapper message rather than a simple float scalar so that it
	// is
	// possible to distinguish between a default value and the value being
	// unset.
	// If omitted, this color object is to be rendered as a solid color
	// (as if the alpha value had been explicitly given with a value of
	// 1.0).
	Alpha float64 `json:"alpha,omitempty"`

	// Blue: The amount of blue in the color as a value in the interval [0,
	// 1].
	Blue float64 `json:"blue,omitempty"`

	// Green: The amount of green in the color as a value in the interval
	// [0, 1].
	Green float64 `json:"green,omitempty"`

	// Red: The amount of red in the color as a value in the interval [0,
	// 1].
	Red float64 `json:"red,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Alpha") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Alpha") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Color) MarshalJSON() ([]byte, error) {
	type NoMethod Color
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *Color) UnmarshalJSON(data []byte) error {
	type NoMethod Color
	var s1 struct {
		Alpha gensupport.JSONFloat64 `json:"alpha"`
		Blue  gensupport.JSONFloat64 `json:"blue"`
		Green gensupport.JSONFloat64 `json:"green"`
		Red   gensupport.JSONFloat64 `json:"red"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Alpha = float64(s1.Alpha)
	s.Blue = float64(s1.Blue)
	s.Green = float64(s1.Green)
	s.Red = float64(s1.Red)
	return nil
}

// ColorInfo: Color information consists of RGB channels, score, and the
// fraction of
// the image that the color occupies in the image.
type ColorInfo struct {
	// Color: RGB components of the color.
	Color *Color `json:"color,omitempty"`

	// PixelFraction: The fraction of pixels the color occupies in the
	// image.
	// Value in range [0, 1].
	PixelFraction float64 `json:"pixelFraction,omitempty"`

	// Score: Image-specific score for this color. Value in range [0, 1].
	Score float64 `json:"score,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Color") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Color") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ColorInfo) MarshalJSON() ([]byte, error) {
	type NoMethod ColorInfo
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *ColorInfo) UnmarshalJSON(data []byte) error {
	type NoMethod ColorInfo
	var s1 struct {
		PixelFraction gensupport.JSONFloat64 `json:"pixelFraction"`
		Score         gensupport.JSONFloat64 `json:"score"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.PixelFraction = float64(s1.PixelFraction)
	s.Score = float64(s1.Score)
	return nil
}

// CropHint: Single crop hint that is used to generate a new crop when
// serving an image.
type CropHint struct {
	// BoundingPoly: The bounding polygon for the crop region. The
	// coordinates of the bounding
	// box are in the original image's scale, as returned in `ImageParams`.
	BoundingPoly *BoundingPoly `json:"boundingPoly,omitempty"`

	// Confidence: Confidence of this being a salient region.  Range [0, 1].
	Confidence float64 `json:"confidence,omitempty"`

	// ImportanceFraction: Fraction of importance of this salient region
	// with respect to the original
	// image.
	ImportanceFraction float64 `json:"importanceFraction,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BoundingPoly") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BoundingPoly") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *CropHint) MarshalJSON() ([]byte, error) {
	type NoMethod CropHint
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *CropHint) UnmarshalJSON(data []byte) error {
	type NoMethod CropHint
	var s1 struct {
		Confidence         gensupport.JSONFloat64 `json:"confidence"`
		ImportanceFraction gensupport.JSONFloat64 `json:"importanceFraction"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	s.ImportanceFraction = float64(s1.ImportanceFraction)
	return nil
}

// CropHintsAnnotation: Set of crop hints that are used to generate new
// crops when serving images.
type CropHintsAnnotation struct {
	// CropHints: Crop hint results.
	CropHints []*CropHint `json:"cropHints,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CropHints") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "CropHints") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *CropHintsAnnotation) MarshalJSON() ([]byte, error) {
	type NoMethod CropHintsAnnotation
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// DetectedBreak: Detected start or end of a structural component.
type DetectedBreak struct {
	// IsPrefix: True if break prepends the element.
	IsPrefix bool `json:"isPrefix,omitempty"`

	// Type: Detected break type.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown break label type.
	//   "SPACE" - Regular space.
	//   "SURE_SPACE" - Sure space (very wide).
	//   "EOL_SURE_SPACE" - Line-wrapping break.
	//   "HYPHEN" - End-line hyphen that is not present in text; does not
	// co-occur with
	// `SPACE`, `LEADER_SPACE`, or `LINE_BREAK`.
	//   "LINE_BREAK" - Line break that ends a paragraph.
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "IsPrefix") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "IsPrefix") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *DetectedBreak) MarshalJSON() ([]byte, error) {
	type NoMethod DetectedBreak
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// DetectedLanguage: Detected language for a structural component.
type DetectedLanguage struct {
	// Confidence: Confidence of detected language. Range [0, 1].
	Confidence float64 `json:"confidence,omitempty"`

	// LanguageCode: The BCP-47 language code, such as "en-US" or "sr-Latn".
	// For more
	// information,
	// see
	// http://www.unicode.org/reports/tr35/#Unicode_locale_identifier.
	LanguageCode string `json:"languageCode,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Confidence") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Confidence") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *DetectedLanguage) MarshalJSON() ([]byte, error) {
	type NoMethod DetectedLanguage
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *DetectedLanguage) UnmarshalJSON(data []byte) error {
	type NoMethod DetectedLanguage
	var s1 struct {
		Confidence gensupport.JSONFloat64 `json:"confidence"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	return nil
}

// DominantColorsAnnotation: Set of dominant colors and their
// corresponding scores.
type DominantColorsAnnotation struct {
	// Colors: RGB color values with their score and pixel fraction.
	Colors []*ColorInfo `json:"colors,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Colors") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Colors") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *DominantColorsAnnotation) MarshalJSON() ([]byte, error) {
	type NoMethod DominantColorsAnnotation
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// EntityAnnotation: Set of detected entity features.
type EntityAnnotation struct {
	// BoundingPoly: Image region to which this entity belongs. Not
	// produced
	// for `LABEL_DETECTION` features.
	BoundingPoly *BoundingPoly `json:"boundingPoly,omitempty"`

	// Confidence: **Deprecated. Use `score` instead.**
	// The accuracy of the entity detection in an image.
	// For example, for an image in which the "Eiffel Tower" entity is
	// detected,
	// this field represents the confidence that there is a tower in the
	// query
	// image. Range [0, 1].
	Confidence float64 `json:"confidence,omitempty"`

	// Description: Entity textual description, expressed in its `locale`
	// language.
	Description string `json:"description,omitempty"`

	// Locale: The language code for the locale in which the entity
	// textual
	// `description` is expressed.
	Locale string `json:"locale,omitempty"`

	// Locations: The location information for the detected entity.
	// Multiple
	// `LocationInfo` elements can be present because one location
	// may
	// indicate the location of the scene in the image, and another
	// location
	// may indicate the location of the place where the image was
	// taken.
	// Location information is usually present for landmarks.
	Locations []*LocationInfo `json:"locations,omitempty"`

	// Mid: Opaque entity ID. Some IDs may be available in
	// [Google Knowledge Graph
	// Search
	// API](https://developers.google.com/knowledge-graph/).
	Mid string `json:"mid,omitempty"`

	// Properties: Some entities may have optional user-supplied `Property`
	// (name/value)
	// fields, such a score or string that qualifies the entity.
	Properties []*Property `json:"properties,omitempty"`

	// Score: Overall score of the result. Range [0, 1].
	Score float64 `json:"score,omitempty"`

	// Topicality: The relevancy of the ICA (Image Content Annotation) label
	// to the
	// image. For example, the relevancy of "tower" is likely higher to an
	// image
	// containing the detected "Eiffel Tower" than to an image containing
	// a
	// detected distant towering building, even though the confidence
	// that
	// there is a tower in each image may be the same. Range [0, 1].
	Topicality float64 `json:"topicality,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BoundingPoly") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BoundingPoly") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *EntityAnnotation) MarshalJSON() ([]byte, error) {
	type NoMethod EntityAnnotation
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *EntityAnnotation) UnmarshalJSON(data []byte) error {
	type NoMethod EntityAnnotation
	var s1 struct {
		Confidence gensupport.JSONFloat64 `json:"confidence"`
		Score      gensupport.JSONFloat64 `json:"score"`
		Topicality gensupport.JSONFloat64 `json:"topicality"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	s.Score = float64(s1.Score)
	s.Topicality = float64(s1.Topicality)
	return nil
}

// FaceAnnotation: A face annotation object contains the results of face
// detection.
type FaceAnnotation struct {
	// AngerLikelihood: Anger likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	AngerLikelihood string `json:"angerLikelihood,omitempty"`

	// BlurredLikelihood: Blurred likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	BlurredLikelihood string `json:"blurredLikelihood,omitempty"`

	// BoundingPoly: The bounding polygon around the face. The coordinates
	// of the bounding box
	// are in the original image's scale, as returned in `ImageParams`.
	// The bounding box is computed to "frame" the face in accordance with
	// human
	// expectations. It is based on the landmarker results.
	// Note that one or more x and/or y coordinates may not be generated in
	// the
	// `BoundingPoly` (the polygon will be unbounded) if only a partial
	// face
	// appears in the image to be annotated.
	BoundingPoly *BoundingPoly `json:"boundingPoly,omitempty"`

	// DetectionConfidence: Detection confidence. Range [0, 1].
	DetectionConfidence float64 `json:"detectionConfidence,omitempty"`

	// FdBoundingPoly: The `fd_bounding_poly` bounding polygon is tighter
	// than the
	// `boundingPoly`, and encloses only the skin part of the face.
	// Typically, it
	// is used to eliminate the face from any image analysis that detects
	// the
	// "amount of skin" visible in an image. It is not based on
	// the
	// landmarker results, only on the initial face detection, hence
	// the <code>fd</code> (face detection) prefix.
	FdBoundingPoly *BoundingPoly `json:"fdBoundingPoly,omitempty"`

	// HeadwearLikelihood: Headwear likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	HeadwearLikelihood string `json:"headwearLikelihood,omitempty"`

	// JoyLikelihood: Joy likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	JoyLikelihood string `json:"joyLikelihood,omitempty"`

	// LandmarkingConfidence: Face landmarking confidence. Range [0, 1].
	LandmarkingConfidence float64 `json:"landmarkingConfidence,omitempty"`

	// Landmarks: Detected face landmarks.
	Landmarks []*Landmark `json:"landmarks,omitempty"`

	// PanAngle: Yaw angle, which indicates the leftward/rightward angle
	// that the face is
	// pointing relative to the vertical plane perpendicular to the image.
	// Range
	// [-180,180].
	PanAngle float64 `json:"panAngle,omitempty"`

	// RollAngle: Roll angle, which indicates the amount of
	// clockwise/anti-clockwise rotation
	// of the face relative to the image vertical about the axis
	// perpendicular to
	// the face. Range [-180,180].
	RollAngle float64 `json:"rollAngle,omitempty"`

	// SorrowLikelihood: Sorrow likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	SorrowLikelihood string `json:"sorrowLikelihood,omitempty"`

	// SurpriseLikelihood: Surprise likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	SurpriseLikelihood string `json:"surpriseLikelihood,omitempty"`

	// TiltAngle: Pitch angle, which indicates the upwards/downwards angle
	// that the face is
	// pointing relative to the image's horizontal plane. Range [-180,180].
	TiltAngle float64 `json:"tiltAngle,omitempty"`

	// UnderExposedLikelihood: Under-exposed likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	UnderExposedLikelihood string `json:"underExposedLikelihood,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AngerLikelihood") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "AngerLikelihood") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *FaceAnnotation) MarshalJSON() ([]byte, error) {
	type NoMethod FaceAnnotation
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *FaceAnnotation) UnmarshalJSON(data []byte) error {
	type NoMethod FaceAnnotation
	var s1 struct {
		DetectionConfidence   gensupport.JSONFloat64 `json:"detectionConfidence"`
		LandmarkingConfidence gensupport.JSONFloat64 `json:"landmarkingConfidence"`
		PanAngle              gensupport.JSONFloat64 `json:"panAngle"`
		RollAngle             gensupport.JSONFloat64 `json:"rollAngle"`
		TiltAngle             gensupport.JSONFloat64 `json:"tiltAngle"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.DetectionConfidence = float64(s1.DetectionConfidence)
	s.LandmarkingConfidence = float64(s1.LandmarkingConfidence)
	s.PanAngle = float64(s1.PanAngle)
	s.RollAngle = float64(s1.RollAngle)
	s.TiltAngle = float64(s1.TiltAngle)
	return nil
}

// GcsDestination: The Google Cloud Storage location where the output
// will be written to.
type GcsDestination struct {
	// Uri: Google Cloud Storage URI where the results will be stored.
	// Results will
	// be in JSON format and preceded by its corresponding input URI. This
	// field
	// can either represent a single file, or a prefix for multiple
	// outputs.
	// Prefixes must end in a `/`.
	//
	// Examples:
	//
	// *    File: gs://bucket-name/filename.json
	// *    Prefix: gs://bucket-name/prefix/here/
	// *    File: gs://bucket-name/prefix/here
	//
	// If multiple outputs, each response is still AnnotateFileResponse,
	// each of
	// which contains some subset of the full list of
	// AnnotateImageResponse.
	// Multiple outputs can happen if, for example, the output JSON is too
	// large
	// and overflows into multiple sharded files.
	Uri string `json:"uri,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Uri") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Uri") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GcsDestination) MarshalJSON() ([]byte, error) {
	type NoMethod GcsDestination
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GcsSource: The Google Cloud Storage location where the input will be
// read from.
type GcsSource struct {
	// Uri: Google Cloud Storage URI for the input file. This must only be
	// a
	// Google Cloud Storage object. Wildcards are not currently supported.
	Uri string `json:"uri,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Uri") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Uri") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GcsSource) MarshalJSON() ([]byte, error) {
	type NoMethod GcsSource
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1AnnotateFileResponse: Response to a single
// file annotation request. A file may contain one or more
// images, which individually have their own responses.
type GoogleCloudVisionV1p2beta1AnnotateFileResponse struct {
	// InputConfig: Information about the file for which this response is
	// generated.
	InputConfig *GoogleCloudVisionV1p2beta1InputConfig `json:"inputConfig,omitempty"`

	// Responses: Individual responses to images found within the file.
	Responses []*GoogleCloudVisionV1p2beta1AnnotateImageResponse `json:"responses,omitempty"`

	// ForceSendFields is a list of field names (e.g. "InputConfig") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "InputConfig") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1AnnotateFileResponse) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1AnnotateFileResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1AnnotateImageRequest: Request for
// performing Google Cloud Vision API tasks over a user-provided
// image, with user-requested features.
type GoogleCloudVisionV1p2beta1AnnotateImageRequest struct {
	// Features: Requested features.
	Features []*GoogleCloudVisionV1p2beta1Feature `json:"features,omitempty"`

	// Image: The image to be processed.
	Image *GoogleCloudVisionV1p2beta1Image `json:"image,omitempty"`

	// ImageContext: Additional context that may accompany the image.
	ImageContext *GoogleCloudVisionV1p2beta1ImageContext `json:"imageContext,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Features") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Features") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1AnnotateImageRequest) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1AnnotateImageRequest
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1AnnotateImageResponse: Response to an image
// annotation request.
type GoogleCloudVisionV1p2beta1AnnotateImageResponse struct {
	// Context: If present, contextual information is needed to understand
	// where this image
	// comes from.
	Context *GoogleCloudVisionV1p2beta1ImageAnnotationContext `json:"context,omitempty"`

	// CropHintsAnnotation: If present, crop hints have completed
	// successfully.
	CropHintsAnnotation *GoogleCloudVisionV1p2beta1CropHintsAnnotation `json:"cropHintsAnnotation,omitempty"`

	// Error: If set, represents the error message for the operation.
	// Note that filled-in image annotations are guaranteed to be
	// correct, even when `error` is set.
	Error *Status `json:"error,omitempty"`

	// FaceAnnotations: If present, face detection has completed
	// successfully.
	FaceAnnotations []*GoogleCloudVisionV1p2beta1FaceAnnotation `json:"faceAnnotations,omitempty"`

	// FullTextAnnotation: If present, text (OCR) detection or document
	// (OCR) text detection has
	// completed successfully.
	// This annotation provides the structural hierarchy for the OCR
	// detected
	// text.
	FullTextAnnotation *GoogleCloudVisionV1p2beta1TextAnnotation `json:"fullTextAnnotation,omitempty"`

	// ImagePropertiesAnnotation: If present, image properties were
	// extracted successfully.
	ImagePropertiesAnnotation *GoogleCloudVisionV1p2beta1ImageProperties `json:"imagePropertiesAnnotation,omitempty"`

	// LabelAnnotations: If present, label detection has completed
	// successfully.
	LabelAnnotations []*GoogleCloudVisionV1p2beta1EntityAnnotation `json:"labelAnnotations,omitempty"`

	// LandmarkAnnotations: If present, landmark detection has completed
	// successfully.
	LandmarkAnnotations []*GoogleCloudVisionV1p2beta1EntityAnnotation `json:"landmarkAnnotations,omitempty"`

	// LogoAnnotations: If present, logo detection has completed
	// successfully.
	LogoAnnotations []*GoogleCloudVisionV1p2beta1EntityAnnotation `json:"logoAnnotations,omitempty"`

	// SafeSearchAnnotation: If present, safe-search annotation has
	// completed successfully.
	SafeSearchAnnotation *GoogleCloudVisionV1p2beta1SafeSearchAnnotation `json:"safeSearchAnnotation,omitempty"`

	// TextAnnotations: If present, text (OCR) detection has completed
	// successfully.
	TextAnnotations []*GoogleCloudVisionV1p2beta1EntityAnnotation `json:"textAnnotations,omitempty"`

	// WebDetection: If present, web detection has completed successfully.
	WebDetection *GoogleCloudVisionV1p2beta1WebDetection `json:"webDetection,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Context") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Context") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1AnnotateImageResponse) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1AnnotateImageResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1AsyncAnnotateFileRequest: An offline file
// annotation request.
type GoogleCloudVisionV1p2beta1AsyncAnnotateFileRequest struct {
	// Features: Required. Requested features.
	Features []*GoogleCloudVisionV1p2beta1Feature `json:"features,omitempty"`

	// ImageContext: Additional context that may accompany the image(s) in
	// the file.
	ImageContext *GoogleCloudVisionV1p2beta1ImageContext `json:"imageContext,omitempty"`

	// InputConfig: Required. Information about the input file.
	InputConfig *GoogleCloudVisionV1p2beta1InputConfig `json:"inputConfig,omitempty"`

	// OutputConfig: Required. The desired output location and metadata
	// (e.g. format).
	OutputConfig *GoogleCloudVisionV1p2beta1OutputConfig `json:"outputConfig,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Features") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Features") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1AsyncAnnotateFileRequest) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1AsyncAnnotateFileRequest
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1AsyncAnnotateFileResponse: The response for
// a single offline file annotation request.
type GoogleCloudVisionV1p2beta1AsyncAnnotateFileResponse struct {
	// OutputConfig: The output location and metadata from
	// AsyncAnnotateFileRequest.
	OutputConfig *GoogleCloudVisionV1p2beta1OutputConfig `json:"outputConfig,omitempty"`

	// ForceSendFields is a list of field names (e.g. "OutputConfig") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "OutputConfig") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1AsyncAnnotateFileResponse) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1AsyncAnnotateFileResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1AsyncBatchAnnotateFilesRequest: Multiple
// async file annotation requests are batched into a single
// service
// call.
type GoogleCloudVisionV1p2beta1AsyncBatchAnnotateFilesRequest struct {
	// Requests: Individual async file annotation requests for this batch.
	Requests []*GoogleCloudVisionV1p2beta1AsyncAnnotateFileRequest `json:"requests,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Requests") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Requests") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1AsyncBatchAnnotateFilesRequest) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1AsyncBatchAnnotateFilesRequest
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1AsyncBatchAnnotateFilesResponse: Response
// to an async batch file annotation request.
type GoogleCloudVisionV1p2beta1AsyncBatchAnnotateFilesResponse struct {
	// Responses: The list of file annotation responses, one for each
	// request in
	// AsyncBatchAnnotateFilesRequest.
	Responses []*GoogleCloudVisionV1p2beta1AsyncAnnotateFileResponse `json:"responses,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Responses") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Responses") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1AsyncBatchAnnotateFilesResponse) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1AsyncBatchAnnotateFilesResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1BatchAnnotateImagesRequest: Multiple image
// annotation requests are batched into a single service call.
type GoogleCloudVisionV1p2beta1BatchAnnotateImagesRequest struct {
	// Requests: Individual image annotation requests for this batch.
	Requests []*GoogleCloudVisionV1p2beta1AnnotateImageRequest `json:"requests,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Requests") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Requests") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1BatchAnnotateImagesRequest) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1BatchAnnotateImagesRequest
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1BatchAnnotateImagesResponse: Response to a
// batch image annotation request.
type GoogleCloudVisionV1p2beta1BatchAnnotateImagesResponse struct {
	// Responses: Individual responses to image annotation requests within
	// the batch.
	Responses []*GoogleCloudVisionV1p2beta1AnnotateImageResponse `json:"responses,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Responses") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Responses") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1BatchAnnotateImagesResponse) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1BatchAnnotateImagesResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1Block: Logical element on the page.
type GoogleCloudVisionV1p2beta1Block struct {
	// BlockType: Detected block type (text, image etc) for this block.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown block type.
	//   "TEXT" - Regular text block.
	//   "TABLE" - Table block.
	//   "PICTURE" - Image block.
	//   "RULER" - Horizontal/vertical line box.
	//   "BARCODE" - Barcode block.
	BlockType string `json:"blockType,omitempty"`

	// BoundingBox: The bounding box for the block.
	// The vertices are in the order of top-left, top-right,
	// bottom-right,
	// bottom-left. When a rotation of the bounding box is detected the
	// rotation
	// is represented as around the top-left corner as defined when the text
	// is
	// read in the 'natural' orientation.
	// For example:
	//
	// * when the text is horizontal it might look like:
	//
	//         0----1
	//         |    |
	//         3----2
	//
	// * when it's rotated 180 degrees around the top-left corner it
	// becomes:
	//
	//         2----3
	//         |    |
	//         1----0
	//
	//   and the vertice order will still be (0, 1, 2, 3).
	BoundingBox *GoogleCloudVisionV1p2beta1BoundingPoly `json:"boundingBox,omitempty"`

	// Confidence: Confidence of the OCR results on the block. Range [0, 1].
	Confidence float64 `json:"confidence,omitempty"`

	// Paragraphs: List of paragraphs in this block (if this blocks is of
	// type text).
	Paragraphs []*GoogleCloudVisionV1p2beta1Paragraph `json:"paragraphs,omitempty"`

	// Property: Additional information detected for the block.
	Property *GoogleCloudVisionV1p2beta1TextAnnotationTextProperty `json:"property,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BlockType") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BlockType") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1Block) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1Block
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1Block) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1Block
	var s1 struct {
		Confidence gensupport.JSONFloat64 `json:"confidence"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	return nil
}

// GoogleCloudVisionV1p2beta1BoundingPoly: A bounding polygon for the
// detected image annotation.
type GoogleCloudVisionV1p2beta1BoundingPoly struct {
	// NormalizedVertices: The bounding polygon normalized vertices.
	NormalizedVertices []*GoogleCloudVisionV1p2beta1NormalizedVertex `json:"normalizedVertices,omitempty"`

	// Vertices: The bounding polygon vertices.
	Vertices []*GoogleCloudVisionV1p2beta1Vertex `json:"vertices,omitempty"`

	// ForceSendFields is a list of field names (e.g. "NormalizedVertices")
	// to unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "NormalizedVertices") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1BoundingPoly) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1BoundingPoly
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1ColorInfo: Color information consists of
// RGB channels, score, and the fraction of
// the image that the color occupies in the image.
type GoogleCloudVisionV1p2beta1ColorInfo struct {
	// Color: RGB components of the color.
	Color *Color `json:"color,omitempty"`

	// PixelFraction: The fraction of pixels the color occupies in the
	// image.
	// Value in range [0, 1].
	PixelFraction float64 `json:"pixelFraction,omitempty"`

	// Score: Image-specific score for this color. Value in range [0, 1].
	Score float64 `json:"score,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Color") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Color") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1ColorInfo) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1ColorInfo
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1ColorInfo) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1ColorInfo
	var s1 struct {
		PixelFraction gensupport.JSONFloat64 `json:"pixelFraction"`
		Score         gensupport.JSONFloat64 `json:"score"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.PixelFraction = float64(s1.PixelFraction)
	s.Score = float64(s1.Score)
	return nil
}

// GoogleCloudVisionV1p2beta1CropHint: Single crop hint that is used to
// generate a new crop when serving an image.
type GoogleCloudVisionV1p2beta1CropHint struct {
	// BoundingPoly: The bounding polygon for the crop region. The
	// coordinates of the bounding
	// box are in the original image's scale, as returned in `ImageParams`.
	BoundingPoly *GoogleCloudVisionV1p2beta1BoundingPoly `json:"boundingPoly,omitempty"`

	// Confidence: Confidence of this being a salient region.  Range [0, 1].
	Confidence float64 `json:"confidence,omitempty"`

	// ImportanceFraction: Fraction of importance of this salient region
	// with respect to the original
	// image.
	ImportanceFraction float64 `json:"importanceFraction,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BoundingPoly") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BoundingPoly") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1CropHint) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1CropHint
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1CropHint) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1CropHint
	var s1 struct {
		Confidence         gensupport.JSONFloat64 `json:"confidence"`
		ImportanceFraction gensupport.JSONFloat64 `json:"importanceFraction"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	s.ImportanceFraction = float64(s1.ImportanceFraction)
	return nil
}

// GoogleCloudVisionV1p2beta1CropHintsAnnotation: Set of crop hints that
// are used to generate new crops when serving images.
type GoogleCloudVisionV1p2beta1CropHintsAnnotation struct {
	// CropHints: Crop hint results.
	CropHints []*GoogleCloudVisionV1p2beta1CropHint `json:"cropHints,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CropHints") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "CropHints") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1CropHintsAnnotation) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1CropHintsAnnotation
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1CropHintsParams: Parameters for crop hints
// annotation request.
type GoogleCloudVisionV1p2beta1CropHintsParams struct {
	// AspectRatios: Aspect ratios in floats, representing the ratio of the
	// width to the height
	// of the image. For example, if the desired aspect ratio is 4/3,
	// the
	// corresponding float value should be 1.33333.  If not specified,
	// the
	// best possible crop is returned. The number of provided aspect ratios
	// is
	// limited to a maximum of 16; any aspect ratios provided after the 16th
	// are
	// ignored.
	AspectRatios []float64 `json:"aspectRatios,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AspectRatios") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "AspectRatios") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1CropHintsParams) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1CropHintsParams
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1DominantColorsAnnotation: Set of dominant
// colors and their corresponding scores.
type GoogleCloudVisionV1p2beta1DominantColorsAnnotation struct {
	// Colors: RGB color values with their score and pixel fraction.
	Colors []*GoogleCloudVisionV1p2beta1ColorInfo `json:"colors,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Colors") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Colors") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1DominantColorsAnnotation) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1DominantColorsAnnotation
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1EntityAnnotation: Set of detected entity
// features.
type GoogleCloudVisionV1p2beta1EntityAnnotation struct {
	// BoundingPoly: Image region to which this entity belongs. Not
	// produced
	// for `LABEL_DETECTION` features.
	BoundingPoly *GoogleCloudVisionV1p2beta1BoundingPoly `json:"boundingPoly,omitempty"`

	// Confidence: **Deprecated. Use `score` instead.**
	// The accuracy of the entity detection in an image.
	// For example, for an image in which the "Eiffel Tower" entity is
	// detected,
	// this field represents the confidence that there is a tower in the
	// query
	// image. Range [0, 1].
	Confidence float64 `json:"confidence,omitempty"`

	// Description: Entity textual description, expressed in its `locale`
	// language.
	Description string `json:"description,omitempty"`

	// Locale: The language code for the locale in which the entity
	// textual
	// `description` is expressed.
	Locale string `json:"locale,omitempty"`

	// Locations: The location information for the detected entity.
	// Multiple
	// `LocationInfo` elements can be present because one location
	// may
	// indicate the location of the scene in the image, and another
	// location
	// may indicate the location of the place where the image was
	// taken.
	// Location information is usually present for landmarks.
	Locations []*GoogleCloudVisionV1p2beta1LocationInfo `json:"locations,omitempty"`

	// Mid: Opaque entity ID. Some IDs may be available in
	// [Google Knowledge Graph
	// Search
	// API](https://developers.google.com/knowledge-graph/).
	Mid string `json:"mid,omitempty"`

	// Properties: Some entities may have optional user-supplied `Property`
	// (name/value)
	// fields, such a score or string that qualifies the entity.
	Properties []*GoogleCloudVisionV1p2beta1Property `json:"properties,omitempty"`

	// Score: Overall score of the result. Range [0, 1].
	Score float64 `json:"score,omitempty"`

	// Topicality: The relevancy of the ICA (Image Content Annotation) label
	// to the
	// image. For example, the relevancy of "tower" is likely higher to an
	// image
	// containing the detected "Eiffel Tower" than to an image containing
	// a
	// detected distant towering building, even though the confidence
	// that
	// there is a tower in each image may be the same. Range [0, 1].
	Topicality float64 `json:"topicality,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BoundingPoly") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BoundingPoly") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1EntityAnnotation) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1EntityAnnotation
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1EntityAnnotation) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1EntityAnnotation
	var s1 struct {
		Confidence gensupport.JSONFloat64 `json:"confidence"`
		Score      gensupport.JSONFloat64 `json:"score"`
		Topicality gensupport.JSONFloat64 `json:"topicality"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	s.Score = float64(s1.Score)
	s.Topicality = float64(s1.Topicality)
	return nil
}

// GoogleCloudVisionV1p2beta1FaceAnnotation: A face annotation object
// contains the results of face detection.
type GoogleCloudVisionV1p2beta1FaceAnnotation struct {
	// AngerLikelihood: Anger likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	AngerLikelihood string `json:"angerLikelihood,omitempty"`

	// BlurredLikelihood: Blurred likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	BlurredLikelihood string `json:"blurredLikelihood,omitempty"`

	// BoundingPoly: The bounding polygon around the face. The coordinates
	// of the bounding box
	// are in the original image's scale, as returned in `ImageParams`.
	// The bounding box is computed to "frame" the face in accordance with
	// human
	// expectations. It is based on the landmarker results.
	// Note that one or more x and/or y coordinates may not be generated in
	// the
	// `BoundingPoly` (the polygon will be unbounded) if only a partial
	// face
	// appears in the image to be annotated.
	BoundingPoly *GoogleCloudVisionV1p2beta1BoundingPoly `json:"boundingPoly,omitempty"`

	// DetectionConfidence: Detection confidence. Range [0, 1].
	DetectionConfidence float64 `json:"detectionConfidence,omitempty"`

	// FdBoundingPoly: The `fd_bounding_poly` bounding polygon is tighter
	// than the
	// `boundingPoly`, and encloses only the skin part of the face.
	// Typically, it
	// is used to eliminate the face from any image analysis that detects
	// the
	// "amount of skin" visible in an image. It is not based on
	// the
	// landmarker results, only on the initial face detection, hence
	// the <code>fd</code> (face detection) prefix.
	FdBoundingPoly *GoogleCloudVisionV1p2beta1BoundingPoly `json:"fdBoundingPoly,omitempty"`

	// HeadwearLikelihood: Headwear likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	HeadwearLikelihood string `json:"headwearLikelihood,omitempty"`

	// JoyLikelihood: Joy likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	JoyLikelihood string `json:"joyLikelihood,omitempty"`

	// LandmarkingConfidence: Face landmarking confidence. Range [0, 1].
	LandmarkingConfidence float64 `json:"landmarkingConfidence,omitempty"`

	// Landmarks: Detected face landmarks.
	Landmarks []*GoogleCloudVisionV1p2beta1FaceAnnotationLandmark `json:"landmarks,omitempty"`

	// PanAngle: Yaw angle, which indicates the leftward/rightward angle
	// that the face is
	// pointing relative to the vertical plane perpendicular to the image.
	// Range
	// [-180,180].
	PanAngle float64 `json:"panAngle,omitempty"`

	// RollAngle: Roll angle, which indicates the amount of
	// clockwise/anti-clockwise rotation
	// of the face relative to the image vertical about the axis
	// perpendicular to
	// the face. Range [-180,180].
	RollAngle float64 `json:"rollAngle,omitempty"`

	// SorrowLikelihood: Sorrow likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	SorrowLikelihood string `json:"sorrowLikelihood,omitempty"`

	// SurpriseLikelihood: Surprise likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	SurpriseLikelihood string `json:"surpriseLikelihood,omitempty"`

	// TiltAngle: Pitch angle, which indicates the upwards/downwards angle
	// that the face is
	// pointing relative to the image's horizontal plane. Range [-180,180].
	TiltAngle float64 `json:"tiltAngle,omitempty"`

	// UnderExposedLikelihood: Under-exposed likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	UnderExposedLikelihood string `json:"underExposedLikelihood,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AngerLikelihood") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "AngerLikelihood") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1FaceAnnotation) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1FaceAnnotation
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1FaceAnnotation) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1FaceAnnotation
	var s1 struct {
		DetectionConfidence   gensupport.JSONFloat64 `json:"detectionConfidence"`
		LandmarkingConfidence gensupport.JSONFloat64 `json:"landmarkingConfidence"`
		PanAngle              gensupport.JSONFloat64 `json:"panAngle"`
		RollAngle             gensupport.JSONFloat64 `json:"rollAngle"`
		TiltAngle             gensupport.JSONFloat64 `json:"tiltAngle"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.DetectionConfidence = float64(s1.DetectionConfidence)
	s.LandmarkingConfidence = float64(s1.LandmarkingConfidence)
	s.PanAngle = float64(s1.PanAngle)
	s.RollAngle = float64(s1.RollAngle)
	s.TiltAngle = float64(s1.TiltAngle)
	return nil
}

// GoogleCloudVisionV1p2beta1FaceAnnotationLandmark: A face-specific
// landmark (for example, a face feature).
type GoogleCloudVisionV1p2beta1FaceAnnotationLandmark struct {
	// Position: Face landmark position.
	Position *GoogleCloudVisionV1p2beta1Position `json:"position,omitempty"`

	// Type: Face landmark type.
	//
	// Possible values:
	//   "UNKNOWN_LANDMARK" - Unknown face landmark detected. Should not be
	// filled.
	//   "LEFT_EYE" - Left eye.
	//   "RIGHT_EYE" - Right eye.
	//   "LEFT_OF_LEFT_EYEBROW" - Left of left eyebrow.
	//   "RIGHT_OF_LEFT_EYEBROW" - Right of left eyebrow.
	//   "LEFT_OF_RIGHT_EYEBROW" - Left of right eyebrow.
	//   "RIGHT_OF_RIGHT_EYEBROW" - Right of right eyebrow.
	//   "MIDPOINT_BETWEEN_EYES" - Midpoint between eyes.
	//   "NOSE_TIP" - Nose tip.
	//   "UPPER_LIP" - Upper lip.
	//   "LOWER_LIP" - Lower lip.
	//   "MOUTH_LEFT" - Mouth left.
	//   "MOUTH_RIGHT" - Mouth right.
	//   "MOUTH_CENTER" - Mouth center.
	//   "NOSE_BOTTOM_RIGHT" - Nose, bottom right.
	//   "NOSE_BOTTOM_LEFT" - Nose, bottom left.
	//   "NOSE_BOTTOM_CENTER" - Nose, bottom center.
	//   "LEFT_EYE_TOP_BOUNDARY" - Left eye, top boundary.
	//   "LEFT_EYE_RIGHT_CORNER" - Left eye, right corner.
	//   "LEFT_EYE_BOTTOM_BOUNDARY" - Left eye, bottom boundary.
	//   "LEFT_EYE_LEFT_CORNER" - Left eye, left corner.
	//   "RIGHT_EYE_TOP_BOUNDARY" - Right eye, top boundary.
	//   "RIGHT_EYE_RIGHT_CORNER" - Right eye, right corner.
	//   "RIGHT_EYE_BOTTOM_BOUNDARY" - Right eye, bottom boundary.
	//   "RIGHT_EYE_LEFT_CORNER" - Right eye, left corner.
	//   "LEFT_EYEBROW_UPPER_MIDPOINT" - Left eyebrow, upper midpoint.
	//   "RIGHT_EYEBROW_UPPER_MIDPOINT" - Right eyebrow, upper midpoint.
	//   "LEFT_EAR_TRAGION" - Left ear tragion.
	//   "RIGHT_EAR_TRAGION" - Right ear tragion.
	//   "LEFT_EYE_PUPIL" - Left eye pupil.
	//   "RIGHT_EYE_PUPIL" - Right eye pupil.
	//   "FOREHEAD_GLABELLA" - Forehead glabella.
	//   "CHIN_GNATHION" - Chin gnathion.
	//   "CHIN_LEFT_GONION" - Chin left gonion.
	//   "CHIN_RIGHT_GONION" - Chin right gonion.
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Position") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Position") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1FaceAnnotationLandmark) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1FaceAnnotationLandmark
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1Feature: The type of Google Cloud Vision
// API detection to perform, and the maximum
// number of results to return for that type. Multiple `Feature` objects
// can
// be specified in the `features` list.
type GoogleCloudVisionV1p2beta1Feature struct {
	// MaxResults: Maximum number of results of this type. Does not apply
	// to
	// `TEXT_DETECTION`, `DOCUMENT_TEXT_DETECTION`, or `CROP_HINTS`.
	MaxResults int64 `json:"maxResults,omitempty"`

	// Model: Model to use for the feature.
	// Supported values: "builtin/stable" (the default if unset)
	// and
	// "builtin/latest".
	Model string `json:"model,omitempty"`

	// Type: The feature type.
	//
	// Possible values:
	//   "TYPE_UNSPECIFIED" - Unspecified feature type.
	//   "FACE_DETECTION" - Run face detection.
	//   "LANDMARK_DETECTION" - Run landmark detection.
	//   "LOGO_DETECTION" - Run logo detection.
	//   "LABEL_DETECTION" - Run label detection.
	//   "TEXT_DETECTION" - Run text detection / optical character
	// recognition (OCR). Text detection
	// is optimized for areas of text within a larger image; if the image
	// is
	// a document, use `DOCUMENT_TEXT_DETECTION` instead.
	//   "DOCUMENT_TEXT_DETECTION" - Run dense text document OCR. Takes
	// precedence when both
	// `DOCUMENT_TEXT_DETECTION` and `TEXT_DETECTION` are present.
	//   "SAFE_SEARCH_DETECTION" - Run Safe Search to detect potentially
	// unsafe
	// or undesirable content.
	//   "IMAGE_PROPERTIES" - Compute a set of image properties, such as
	// the
	// image's dominant colors.
	//   "CROP_HINTS" - Run crop hints.
	//   "WEB_DETECTION" - Run web detection.
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "MaxResults") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "MaxResults") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1Feature) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1Feature
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1GcsDestination: The Google Cloud Storage
// location where the output will be written to.
type GoogleCloudVisionV1p2beta1GcsDestination struct {
	// Uri: Google Cloud Storage URI where the results will be stored.
	// Results will
	// be in JSON format and preceded by its corresponding input URI. This
	// field
	// can either represent a single file, or a prefix for multiple
	// outputs.
	// Prefixes must end in a `/`.
	//
	// Examples:
	//
	// *    File: gs://bucket-name/filename.json
	// *    Prefix: gs://bucket-name/prefix/here/
	// *    File: gs://bucket-name/prefix/here
	//
	// If multiple outputs, each response is still AnnotateFileResponse,
	// each of
	// which contains some subset of the full list of
	// AnnotateImageResponse.
	// Multiple outputs can happen if, for example, the output JSON is too
	// large
	// and overflows into multiple sharded files.
	Uri string `json:"uri,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Uri") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Uri") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1GcsDestination) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1GcsDestination
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1GcsSource: The Google Cloud Storage
// location where the input will be read from.
type GoogleCloudVisionV1p2beta1GcsSource struct {
	// Uri: Google Cloud Storage URI for the input file. This must only be
	// a
	// Google Cloud Storage object. Wildcards are not currently supported.
	Uri string `json:"uri,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Uri") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Uri") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1GcsSource) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1GcsSource
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1Image: Client image to perform Google Cloud
// Vision API tasks over.
type GoogleCloudVisionV1p2beta1Image struct {
	// Content: Image content, represented as a stream of bytes.
	// Note: As with all `bytes` fields, protobuffers use a pure
	// binary
	// representation, whereas JSON representations use base64.
	Content string `json:"content,omitempty"`

	// Source: Google Cloud Storage image location, or publicly-accessible
	// image
	// URL. If both `content` and `source` are provided for an image,
	// `content`
	// takes precedence and is used to perform the image annotation request.
	Source *GoogleCloudVisionV1p2beta1ImageSource `json:"source,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Content") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Content") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1Image) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1Image
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1ImageAnnotationContext: If an image was
// produced from a file (e.g. a PDF), this message gives
// information about the source of that image.
type GoogleCloudVisionV1p2beta1ImageAnnotationContext struct {
	// PageNumber: If the file was a PDF or TIFF, this field gives the page
	// number within
	// the file used to produce the image.
	PageNumber int64 `json:"pageNumber,omitempty"`

	// Uri: The URI of the file used to produce the image.
	Uri string `json:"uri,omitempty"`

	// ForceSendFields is a list of field names (e.g. "PageNumber") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "PageNumber") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1ImageAnnotationContext) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1ImageAnnotationContext
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1ImageContext: Image context and/or
// feature-specific parameters.
type GoogleCloudVisionV1p2beta1ImageContext struct {
	// CropHintsParams: Parameters for crop hints annotation request.
	CropHintsParams *GoogleCloudVisionV1p2beta1CropHintsParams `json:"cropHintsParams,omitempty"`

	// LanguageHints: List of languages to use for TEXT_DETECTION. In most
	// cases, an empty value
	// yields the best results since it enables automatic language
	// detection. For
	// languages based on the Latin alphabet, setting `language_hints` is
	// not
	// needed. In rare cases, when the language of the text in the image is
	// known,
	// setting a hint will help get better results (although it will be
	// a
	// significant hindrance if the hint is wrong). Text detection returns
	// an
	// error if one or more of the specified languages is not one of
	// the
	// [supported languages](/vision/docs/languages).
	LanguageHints []string `json:"languageHints,omitempty"`

	// LatLongRect: Not used.
	LatLongRect *GoogleCloudVisionV1p2beta1LatLongRect `json:"latLongRect,omitempty"`

	// WebDetectionParams: Parameters for web detection.
	WebDetectionParams *GoogleCloudVisionV1p2beta1WebDetectionParams `json:"webDetectionParams,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CropHintsParams") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "CropHintsParams") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1ImageContext) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1ImageContext
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1ImageProperties: Stores image properties,
// such as dominant colors.
type GoogleCloudVisionV1p2beta1ImageProperties struct {
	// DominantColors: If present, dominant colors completed successfully.
	DominantColors *GoogleCloudVisionV1p2beta1DominantColorsAnnotation `json:"dominantColors,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DominantColors") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "DominantColors") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1ImageProperties) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1ImageProperties
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1ImageSource: External image source (Google
// Cloud Storage or web URL image location).
type GoogleCloudVisionV1p2beta1ImageSource struct {
	// GcsImageUri: **Use `image_uri` instead.**
	//
	// The Google Cloud Storage  URI of the
	// form
	// `gs://bucket_name/object_name`. Object versioning is not supported.
	// See
	// [Google Cloud Storage
	// Request
	// URIs](https://cloud.google.com/storage/docs/reference-uris) for more
	// info.
	GcsImageUri string `json:"gcsImageUri,omitempty"`

	// ImageUri: The URI of the source image. Can be either:
	//
	// 1. A Google Cloud Storage URI of the form
	//    `gs://bucket_name/object_name`. Object versioning is not
	// supported. See
	//    [Google Cloud Storage Request
	//    URIs](https://cloud.google.com/storage/docs/reference-uris) for
	// more
	//    info.
	//
	// 2. A publicly-accessible image HTTP/HTTPS URL. When fetching images
	// from
	//    HTTP/HTTPS URLs, Google cannot guarantee that the request will be
	//    completed. Your request may fail if the specified host denies the
	//    request (e.g. due to request throttling or DOS prevention), or if
	// Google
	//    throttles requests to the site for abuse prevention. You should
	// not
	//    depend on externally-hosted images for production
	// applications.
	//
	// When both `gcs_image_uri` and `image_uri` are specified, `image_uri`
	// takes
	// precedence.
	ImageUri string `json:"imageUri,omitempty"`

	// ForceSendFields is a list of field names (e.g. "GcsImageUri") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "GcsImageUri") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1ImageSource) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1ImageSource
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1InputConfig: The desired input location and
// metadata.
type GoogleCloudVisionV1p2beta1InputConfig struct {
	// GcsSource: The Google Cloud Storage location to read the input from.
	GcsSource *GoogleCloudVisionV1p2beta1GcsSource `json:"gcsSource,omitempty"`

	// MimeType: The type of the file. Currently only "application/pdf" and
	// "image/tiff"
	// are supported. Wildcards are not supported.
	MimeType string `json:"mimeType,omitempty"`

	// ForceSendFields is a list of field names (e.g. "GcsSource") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "GcsSource") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1InputConfig) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1InputConfig
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1LatLongRect: Rectangle determined by min
// and max `LatLng` pairs.
type GoogleCloudVisionV1p2beta1LatLongRect struct {
	// MaxLatLng: Max lat/long pair.
	MaxLatLng *LatLng `json:"maxLatLng,omitempty"`

	// MinLatLng: Min lat/long pair.
	MinLatLng *LatLng `json:"minLatLng,omitempty"`

	// ForceSendFields is a list of field names (e.g. "MaxLatLng") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "MaxLatLng") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1LatLongRect) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1LatLongRect
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1LocationInfo: Detected entity location
// information.
type GoogleCloudVisionV1p2beta1LocationInfo struct {
	// LatLng: lat/long location coordinates.
	LatLng *LatLng `json:"latLng,omitempty"`

	// ForceSendFields is a list of field names (e.g. "LatLng") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "LatLng") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1LocationInfo) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1LocationInfo
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1NormalizedVertex: A vertex represents a 2D
// point in the image.
// NOTE: the normalized vertex coordinates are relative to the original
// image
// and range from 0 to 1.
type GoogleCloudVisionV1p2beta1NormalizedVertex struct {
	// X: X coordinate.
	X float64 `json:"x,omitempty"`

	// Y: Y coordinate.
	Y float64 `json:"y,omitempty"`

	// ForceSendFields is a list of field names (e.g. "X") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "X") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1NormalizedVertex) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1NormalizedVertex
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1NormalizedVertex) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1NormalizedVertex
	var s1 struct {
		X gensupport.JSONFloat64 `json:"x"`
		Y gensupport.JSONFloat64 `json:"y"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.X = float64(s1.X)
	s.Y = float64(s1.Y)
	return nil
}

// GoogleCloudVisionV1p2beta1OperationMetadata: Contains metadata for
// the BatchAnnotateImages operation.
type GoogleCloudVisionV1p2beta1OperationMetadata struct {
	// CreateTime: The time when the batch request was received.
	CreateTime string `json:"createTime,omitempty"`

	// State: Current state of the batch operation.
	//
	// Possible values:
	//   "STATE_UNSPECIFIED" - Invalid.
	//   "CREATED" - Request is received.
	//   "RUNNING" - Request is actively being processed.
	//   "DONE" - The batch processing is done.
	//   "CANCELLED" - The batch processing was cancelled.
	State string `json:"state,omitempty"`

	// UpdateTime: The time when the operation result was last updated.
	UpdateTime string `json:"updateTime,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CreateTime") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "CreateTime") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1OperationMetadata) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1OperationMetadata
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1OutputConfig: The desired output location
// and metadata.
type GoogleCloudVisionV1p2beta1OutputConfig struct {
	// BatchSize: The max number of response protos to put into each output
	// JSON file on
	// Google Cloud Storage.
	// The valid range is [1, 100]. If not specified, the default value is
	// 20.
	//
	// For example, for one pdf file with 100 pages, 100 response protos
	// will
	// be generated. If `batch_size` = 20, then 5 json files each
	// containing 20 response protos will be written under the
	// prefix
	// `gcs_destination`.`uri`.
	//
	// Currently, batch_size only applies to GcsDestination, with potential
	// future
	// support for other output configurations.
	BatchSize int64 `json:"batchSize,omitempty"`

	// GcsDestination: The Google Cloud Storage location to write the
	// output(s) to.
	GcsDestination *GoogleCloudVisionV1p2beta1GcsDestination `json:"gcsDestination,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BatchSize") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BatchSize") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1OutputConfig) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1OutputConfig
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1Page: Detected page from OCR.
type GoogleCloudVisionV1p2beta1Page struct {
	// Blocks: List of blocks of text, images etc on this page.
	Blocks []*GoogleCloudVisionV1p2beta1Block `json:"blocks,omitempty"`

	// Confidence: Confidence of the OCR results on the page. Range [0, 1].
	Confidence float64 `json:"confidence,omitempty"`

	// Height: Page height. For PDFs the unit is points. For images
	// (including
	// TIFFs) the unit is pixels.
	Height int64 `json:"height,omitempty"`

	// Property: Additional information detected on the page.
	Property *GoogleCloudVisionV1p2beta1TextAnnotationTextProperty `json:"property,omitempty"`

	// Width: Page width. For PDFs the unit is points. For images
	// (including
	// TIFFs) the unit is pixels.
	Width int64 `json:"width,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Blocks") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Blocks") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1Page) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1Page
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1Page) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1Page
	var s1 struct {
		Confidence gensupport.JSONFloat64 `json:"confidence"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	return nil
}

// GoogleCloudVisionV1p2beta1Paragraph: Structural unit of text
// representing a number of words in certain order.
type GoogleCloudVisionV1p2beta1Paragraph struct {
	// BoundingBox: The bounding box for the paragraph.
	// The vertices are in the order of top-left, top-right,
	// bottom-right,
	// bottom-left. When a rotation of the bounding box is detected the
	// rotation
	// is represented as around the top-left corner as defined when the text
	// is
	// read in the 'natural' orientation.
	// For example:
	//   * when the text is horizontal it might look like:
	//      0----1
	//      |    |
	//      3----2
	//   * when it's rotated 180 degrees around the top-left corner it
	// becomes:
	//      2----3
	//      |    |
	//      1----0
	//   and the vertice order will still be (0, 1, 2, 3).
	BoundingBox *GoogleCloudVisionV1p2beta1BoundingPoly `json:"boundingBox,omitempty"`

	// Confidence: Confidence of the OCR results for the paragraph. Range
	// [0, 1].
	Confidence float64 `json:"confidence,omitempty"`

	// Property: Additional information detected for the paragraph.
	Property *GoogleCloudVisionV1p2beta1TextAnnotationTextProperty `json:"property,omitempty"`

	// Words: List of words in this paragraph.
	Words []*GoogleCloudVisionV1p2beta1Word `json:"words,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BoundingBox") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BoundingBox") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1Paragraph) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1Paragraph
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1Paragraph) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1Paragraph
	var s1 struct {
		Confidence gensupport.JSONFloat64 `json:"confidence"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	return nil
}

// GoogleCloudVisionV1p2beta1Position: A 3D position in the image, used
// primarily for Face detection landmarks.
// A valid Position must have both x and y coordinates.
// The position coordinates are in the same scale as the original image.
type GoogleCloudVisionV1p2beta1Position struct {
	// X: X coordinate.
	X float64 `json:"x,omitempty"`

	// Y: Y coordinate.
	Y float64 `json:"y,omitempty"`

	// Z: Z coordinate (or depth).
	Z float64 `json:"z,omitempty"`

	// ForceSendFields is a list of field names (e.g. "X") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "X") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1Position) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1Position
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1Position) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1Position
	var s1 struct {
		X gensupport.JSONFloat64 `json:"x"`
		Y gensupport.JSONFloat64 `json:"y"`
		Z gensupport.JSONFloat64 `json:"z"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.X = float64(s1.X)
	s.Y = float64(s1.Y)
	s.Z = float64(s1.Z)
	return nil
}

// GoogleCloudVisionV1p2beta1Property: A `Property` consists of a
// user-supplied name/value pair.
type GoogleCloudVisionV1p2beta1Property struct {
	// Name: Name of the property.
	Name string `json:"name,omitempty"`

	// Uint64Value: Value of numeric properties.
	Uint64Value uint64 `json:"uint64Value,omitempty,string"`

	// Value: Value of the property.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Name") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Name") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1Property) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1Property
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1SafeSearchAnnotation: Set of features
// pertaining to the image, computed by computer vision
// methods over safe-search verticals (for example, adult, spoof,
// medical,
// violence).
type GoogleCloudVisionV1p2beta1SafeSearchAnnotation struct {
	// Adult: Represents the adult content likelihood for the image. Adult
	// content may
	// contain elements such as nudity, pornographic images or cartoons,
	// or
	// sexual activities.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	Adult string `json:"adult,omitempty"`

	// Medical: Likelihood that this is a medical image.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	Medical string `json:"medical,omitempty"`

	// Racy: Likelihood that the request image contains racy content. Racy
	// content may
	// include (but is not limited to) skimpy or sheer clothing,
	// strategically
	// covered nudity, lewd or provocative poses, or close-ups of
	// sensitive
	// body areas.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	Racy string `json:"racy,omitempty"`

	// Spoof: Spoof likelihood. The likelihood that an modification
	// was made to the image's canonical version to make it appear
	// funny or offensive.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	Spoof string `json:"spoof,omitempty"`

	// Violence: Likelihood that this image contains violent content.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	Violence string `json:"violence,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Adult") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Adult") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1SafeSearchAnnotation) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1SafeSearchAnnotation
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1Symbol: A single symbol representation.
type GoogleCloudVisionV1p2beta1Symbol struct {
	// BoundingBox: The bounding box for the symbol.
	// The vertices are in the order of top-left, top-right,
	// bottom-right,
	// bottom-left. When a rotation of the bounding box is detected the
	// rotation
	// is represented as around the top-left corner as defined when the text
	// is
	// read in the 'natural' orientation.
	// For example:
	//   * when the text is horizontal it might look like:
	//      0----1
	//      |    |
	//      3----2
	//   * when it's rotated 180 degrees around the top-left corner it
	// becomes:
	//      2----3
	//      |    |
	//      1----0
	//   and the vertice order will still be (0, 1, 2, 3).
	BoundingBox *GoogleCloudVisionV1p2beta1BoundingPoly `json:"boundingBox,omitempty"`

	// Confidence: Confidence of the OCR results for the symbol. Range [0,
	// 1].
	Confidence float64 `json:"confidence,omitempty"`

	// Property: Additional information detected for the symbol.
	Property *GoogleCloudVisionV1p2beta1TextAnnotationTextProperty `json:"property,omitempty"`

	// Text: The actual UTF-8 representation of the symbol.
	Text string `json:"text,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BoundingBox") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BoundingBox") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1Symbol) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1Symbol
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1Symbol) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1Symbol
	var s1 struct {
		Confidence gensupport.JSONFloat64 `json:"confidence"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	return nil
}

// GoogleCloudVisionV1p2beta1TextAnnotation: TextAnnotation contains a
// structured representation of OCR extracted text.
// The hierarchy of an OCR extracted text structure is like this:
//     TextAnnotation -> Page -> Block -> Paragraph -> Word ->
// Symbol
// Each structural component, starting from Page, may further have their
// own
// properties. Properties describe detected languages, breaks etc..
// Please refer
// to the TextAnnotation.TextProperty message definition below for
// more
// detail.
type GoogleCloudVisionV1p2beta1TextAnnotation struct {
	// Pages: List of pages detected by OCR.
	Pages []*GoogleCloudVisionV1p2beta1Page `json:"pages,omitempty"`

	// Text: UTF-8 text detected on the pages.
	Text string `json:"text,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Pages") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Pages") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1TextAnnotation) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1TextAnnotation
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1TextAnnotationDetectedBreak: Detected start
// or end of a structural component.
type GoogleCloudVisionV1p2beta1TextAnnotationDetectedBreak struct {
	// IsPrefix: True if break prepends the element.
	IsPrefix bool `json:"isPrefix,omitempty"`

	// Type: Detected break type.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown break label type.
	//   "SPACE" - Regular space.
	//   "SURE_SPACE" - Sure space (very wide).
	//   "EOL_SURE_SPACE" - Line-wrapping break.
	//   "HYPHEN" - End-line hyphen that is not present in text; does not
	// co-occur with
	// `SPACE`, `LEADER_SPACE`, or `LINE_BREAK`.
	//   "LINE_BREAK" - Line break that ends a paragraph.
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "IsPrefix") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "IsPrefix") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1TextAnnotationDetectedBreak) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1TextAnnotationDetectedBreak
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1TextAnnotationDetectedLanguage: Detected
// language for a structural component.
type GoogleCloudVisionV1p2beta1TextAnnotationDetectedLanguage struct {
	// Confidence: Confidence of detected language. Range [0, 1].
	Confidence float64 `json:"confidence,omitempty"`

	// LanguageCode: The BCP-47 language code, such as "en-US" or "sr-Latn".
	// For more
	// information,
	// see
	// http://www.unicode.org/reports/tr35/#Unicode_locale_identifier.
	LanguageCode string `json:"languageCode,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Confidence") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Confidence") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1TextAnnotationDetectedLanguage) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1TextAnnotationDetectedLanguage
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1TextAnnotationDetectedLanguage) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1TextAnnotationDetectedLanguage
	var s1 struct {
		Confidence gensupport.JSONFloat64 `json:"confidence"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	return nil
}

// GoogleCloudVisionV1p2beta1TextAnnotationTextProperty: Additional
// information detected on the structural component.
type GoogleCloudVisionV1p2beta1TextAnnotationTextProperty struct {
	// DetectedBreak: Detected start or end of a text segment.
	DetectedBreak *GoogleCloudVisionV1p2beta1TextAnnotationDetectedBreak `json:"detectedBreak,omitempty"`

	// DetectedLanguages: A list of detected languages together with
	// confidence.
	DetectedLanguages []*GoogleCloudVisionV1p2beta1TextAnnotationDetectedLanguage `json:"detectedLanguages,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DetectedBreak") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "DetectedBreak") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1TextAnnotationTextProperty) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1TextAnnotationTextProperty
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1Vertex: A vertex represents a 2D point in
// the image.
// NOTE: the vertex coordinates are in the same scale as the original
// image.
type GoogleCloudVisionV1p2beta1Vertex struct {
	// X: X coordinate.
	X int64 `json:"x,omitempty"`

	// Y: Y coordinate.
	Y int64 `json:"y,omitempty"`

	// ForceSendFields is a list of field names (e.g. "X") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "X") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1Vertex) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1Vertex
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1WebDetection: Relevant information for the
// image from the Internet.
type GoogleCloudVisionV1p2beta1WebDetection struct {
	// BestGuessLabels: The service's best guess as to the topic of the
	// request image.
	// Inferred from similar images on the open web.
	BestGuessLabels []*GoogleCloudVisionV1p2beta1WebDetectionWebLabel `json:"bestGuessLabels,omitempty"`

	// FullMatchingImages: Fully matching images from the Internet.
	// Can include resized copies of the query image.
	FullMatchingImages []*GoogleCloudVisionV1p2beta1WebDetectionWebImage `json:"fullMatchingImages,omitempty"`

	// PagesWithMatchingImages: Web pages containing the matching images
	// from the Internet.
	PagesWithMatchingImages []*GoogleCloudVisionV1p2beta1WebDetectionWebPage `json:"pagesWithMatchingImages,omitempty"`

	// PartialMatchingImages: Partial matching images from the
	// Internet.
	// Those images are similar enough to share some key-point features.
	// For
	// example an original image will likely have partial matching for its
	// crops.
	PartialMatchingImages []*GoogleCloudVisionV1p2beta1WebDetectionWebImage `json:"partialMatchingImages,omitempty"`

	// VisuallySimilarImages: The visually similar image results.
	VisuallySimilarImages []*GoogleCloudVisionV1p2beta1WebDetectionWebImage `json:"visuallySimilarImages,omitempty"`

	// WebEntities: Deduced entities from similar images on the Internet.
	WebEntities []*GoogleCloudVisionV1p2beta1WebDetectionWebEntity `json:"webEntities,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BestGuessLabels") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BestGuessLabels") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1WebDetection) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1WebDetection
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1WebDetectionParams: Parameters for web
// detection request.
type GoogleCloudVisionV1p2beta1WebDetectionParams struct {
	// IncludeGeoResults: Whether to include results derived from the geo
	// information in the image.
	IncludeGeoResults bool `json:"includeGeoResults,omitempty"`

	// ForceSendFields is a list of field names (e.g. "IncludeGeoResults")
	// to unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "IncludeGeoResults") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1WebDetectionParams) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1WebDetectionParams
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1WebDetectionWebEntity: Entity deduced from
// similar images on the Internet.
type GoogleCloudVisionV1p2beta1WebDetectionWebEntity struct {
	// Description: Canonical description of the entity, in English.
	Description string `json:"description,omitempty"`

	// EntityId: Opaque entity ID.
	EntityId string `json:"entityId,omitempty"`

	// Score: Overall relevancy score for the entity.
	// Not normalized and not comparable across different image queries.
	Score float64 `json:"score,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Description") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Description") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1WebDetectionWebEntity) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1WebDetectionWebEntity
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1WebDetectionWebEntity) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1WebDetectionWebEntity
	var s1 struct {
		Score gensupport.JSONFloat64 `json:"score"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Score = float64(s1.Score)
	return nil
}

// GoogleCloudVisionV1p2beta1WebDetectionWebImage: Metadata for online
// images.
type GoogleCloudVisionV1p2beta1WebDetectionWebImage struct {
	// Score: (Deprecated) Overall relevancy score for the image.
	Score float64 `json:"score,omitempty"`

	// Url: The result image URL.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Score") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Score") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1WebDetectionWebImage) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1WebDetectionWebImage
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1WebDetectionWebImage) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1WebDetectionWebImage
	var s1 struct {
		Score gensupport.JSONFloat64 `json:"score"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Score = float64(s1.Score)
	return nil
}

// GoogleCloudVisionV1p2beta1WebDetectionWebLabel: Label to provide
// extra metadata for the web detection.
type GoogleCloudVisionV1p2beta1WebDetectionWebLabel struct {
	// Label: Label for extra metadata.
	Label string `json:"label,omitempty"`

	// LanguageCode: The BCP-47 language code for `label`, such as "en-US"
	// or "sr-Latn".
	// For more information,
	// see
	// http://www.unicode.org/reports/tr35/#Unicode_locale_identifier.
	LanguageCode string `json:"languageCode,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Label") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Label") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1WebDetectionWebLabel) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1WebDetectionWebLabel
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p2beta1WebDetectionWebPage: Metadata for web
// pages.
type GoogleCloudVisionV1p2beta1WebDetectionWebPage struct {
	// FullMatchingImages: Fully matching images on the page.
	// Can include resized copies of the query image.
	FullMatchingImages []*GoogleCloudVisionV1p2beta1WebDetectionWebImage `json:"fullMatchingImages,omitempty"`

	// PageTitle: Title for the web page, may contain HTML markups.
	PageTitle string `json:"pageTitle,omitempty"`

	// PartialMatchingImages: Partial matching images on the page.
	// Those images are similar enough to share some key-point features.
	// For
	// example an original image will likely have partial matching for
	// its
	// crops.
	PartialMatchingImages []*GoogleCloudVisionV1p2beta1WebDetectionWebImage `json:"partialMatchingImages,omitempty"`

	// Score: (Deprecated) Overall relevancy score for the web page.
	Score float64 `json:"score,omitempty"`

	// Url: The result web page URL.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FullMatchingImages")
	// to unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "FullMatchingImages") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1WebDetectionWebPage) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1WebDetectionWebPage
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1WebDetectionWebPage) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1WebDetectionWebPage
	var s1 struct {
		Score gensupport.JSONFloat64 `json:"score"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Score = float64(s1.Score)
	return nil
}

// GoogleCloudVisionV1p2beta1Word: A word representation.
type GoogleCloudVisionV1p2beta1Word struct {
	// BoundingBox: The bounding box for the word.
	// The vertices are in the order of top-left, top-right,
	// bottom-right,
	// bottom-left. When a rotation of the bounding box is detected the
	// rotation
	// is represented as around the top-left corner as defined when the text
	// is
	// read in the 'natural' orientation.
	// For example:
	//   * when the text is horizontal it might look like:
	//      0----1
	//      |    |
	//      3----2
	//   * when it's rotated 180 degrees around the top-left corner it
	// becomes:
	//      2----3
	//      |    |
	//      1----0
	//   and the vertice order will still be (0, 1, 2, 3).
	BoundingBox *GoogleCloudVisionV1p2beta1BoundingPoly `json:"boundingBox,omitempty"`

	// Confidence: Confidence of the OCR results for the word. Range [0, 1].
	Confidence float64 `json:"confidence,omitempty"`

	// Property: Additional information detected for the word.
	Property *GoogleCloudVisionV1p2beta1TextAnnotationTextProperty `json:"property,omitempty"`

	// Symbols: List of symbols in the word.
	// The order of the symbols follows the natural reading order.
	Symbols []*GoogleCloudVisionV1p2beta1Symbol `json:"symbols,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BoundingBox") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BoundingBox") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p2beta1Word) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p2beta1Word
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p2beta1Word) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p2beta1Word
	var s1 struct {
		Confidence gensupport.JSONFloat64 `json:"confidence"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	return nil
}

// GoogleCloudVisionV1p3beta1BatchOperationMetadata: Metadata for the
// batch operations such as the current state.
//
// This is included in the `metadata` field of the `Operation` returned
// by the
// `GetOperation` call of the `google::longrunning::Operations` service.
type GoogleCloudVisionV1p3beta1BatchOperationMetadata struct {
	// EndTime: The time when the batch request is finished
	// and
	// google.longrunning.Operation.done is set to true.
	EndTime string `json:"endTime,omitempty"`

	// State: The current state of the batch operation.
	//
	// Possible values:
	//   "STATE_UNSPECIFIED" - Invalid.
	//   "PROCESSING" - Request is actively being processed.
	//   "SUCCESSFUL" - The request is done and at least one item has been
	// successfully
	// processed.
	//   "FAILED" - The request is done and no item has been successfully
	// processed.
	//   "CANCELLED" - The request is done after the
	// longrunning.Operations.CancelOperation has
	// been called by the user.  Any records that were processed before
	// the
	// cancel command are output as specified in the request.
	State string `json:"state,omitempty"`

	// SubmitTime: The time when the batch request was submitted to the
	// server.
	SubmitTime string `json:"submitTime,omitempty"`

	// ForceSendFields is a list of field names (e.g. "EndTime") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "EndTime") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p3beta1BatchOperationMetadata) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p3beta1BatchOperationMetadata
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p3beta1BoundingPoly: A bounding polygon for the
// detected image annotation.
type GoogleCloudVisionV1p3beta1BoundingPoly struct {
	// NormalizedVertices: The bounding polygon normalized vertices.
	NormalizedVertices []*GoogleCloudVisionV1p3beta1NormalizedVertex `json:"normalizedVertices,omitempty"`

	// Vertices: The bounding polygon vertices.
	Vertices []*GoogleCloudVisionV1p3beta1Vertex `json:"vertices,omitempty"`

	// ForceSendFields is a list of field names (e.g. "NormalizedVertices")
	// to unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "NormalizedVertices") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p3beta1BoundingPoly) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p3beta1BoundingPoly
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p3beta1ImportProductSetsResponse: Response message
// for the `ImportProductSets` method.
//
// This message is returned by
// the
// google.longrunning.Operations.GetOperation method in the
// returned
// google.longrunning.Operation.response field.
type GoogleCloudVisionV1p3beta1ImportProductSetsResponse struct {
	// ReferenceImages: The list of reference_images that are imported
	// successfully.
	ReferenceImages []*GoogleCloudVisionV1p3beta1ReferenceImage `json:"referenceImages,omitempty"`

	// Statuses: The rpc status for each ImportProductSet request, including
	// both successes
	// and errors.
	//
	// The number of statuses here matches the number of lines in the csv
	// file,
	// and statuses[i] stores the success or failure status of processing
	// the i-th
	// line of the csv, starting from line 0.
	Statuses []*Status `json:"statuses,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ReferenceImages") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "ReferenceImages") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p3beta1ImportProductSetsResponse) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p3beta1ImportProductSetsResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p3beta1NormalizedVertex: A vertex represents a 2D
// point in the image.
// NOTE: the normalized vertex coordinates are relative to the original
// image
// and range from 0 to 1.
type GoogleCloudVisionV1p3beta1NormalizedVertex struct {
	// X: X coordinate.
	X float64 `json:"x,omitempty"`

	// Y: Y coordinate.
	Y float64 `json:"y,omitempty"`

	// ForceSendFields is a list of field names (e.g. "X") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "X") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p3beta1NormalizedVertex) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p3beta1NormalizedVertex
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *GoogleCloudVisionV1p3beta1NormalizedVertex) UnmarshalJSON(data []byte) error {
	type NoMethod GoogleCloudVisionV1p3beta1NormalizedVertex
	var s1 struct {
		X gensupport.JSONFloat64 `json:"x"`
		Y gensupport.JSONFloat64 `json:"y"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.X = float64(s1.X)
	s.Y = float64(s1.Y)
	return nil
}

// GoogleCloudVisionV1p3beta1ReferenceImage: A `ReferenceImage`
// represents a product image and its associated metadata,
// such as bounding boxes.
type GoogleCloudVisionV1p3beta1ReferenceImage struct {
	// BoundingPolys: Bounding polygons around the areas of interest in the
	// reference image.
	// Optional. If this field is empty, the system will try to detect
	// regions of
	// interest. At most 10 bounding polygons will be used.
	//
	// The provided shape is converted into a non-rotated rectangle.
	// Once
	// converted, the small edge of the rectangle must be greater than or
	// equal
	// to 300 pixels. The aspect ratio must be 1:4 or less (i.e. 1:3 is ok;
	// 1:5
	// is not).
	BoundingPolys []*GoogleCloudVisionV1p3beta1BoundingPoly `json:"boundingPolys,omitempty"`

	// Name: The resource name of the reference image.
	//
	// Format
	// is:
	//
	// `projects/PROJECT_ID/locations/LOC_ID/products/PRODUCT_ID/referen
	// ceImages/IMAGE_ID`.
	//
	// This field is ignored when creating a reference image.
	Name string `json:"name,omitempty"`

	// Uri: The Google Cloud Storage URI of the reference image.
	//
	// The URI must start with `gs://`.
	//
	// Required.
	Uri string `json:"uri,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BoundingPolys") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BoundingPolys") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p3beta1ReferenceImage) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p3beta1ReferenceImage
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GoogleCloudVisionV1p3beta1Vertex: A vertex represents a 2D point in
// the image.
// NOTE: the vertex coordinates are in the same scale as the original
// image.
type GoogleCloudVisionV1p3beta1Vertex struct {
	// X: X coordinate.
	X int64 `json:"x,omitempty"`

	// Y: Y coordinate.
	Y int64 `json:"y,omitempty"`

	// ForceSendFields is a list of field names (e.g. "X") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "X") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GoogleCloudVisionV1p3beta1Vertex) MarshalJSON() ([]byte, error) {
	type NoMethod GoogleCloudVisionV1p3beta1Vertex
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ImageAnnotationContext: If an image was produced from a file (e.g. a
// PDF), this message gives
// information about the source of that image.
type ImageAnnotationContext struct {
	// PageNumber: If the file was a PDF or TIFF, this field gives the page
	// number within
	// the file used to produce the image.
	PageNumber int64 `json:"pageNumber,omitempty"`

	// Uri: The URI of the file used to produce the image.
	Uri string `json:"uri,omitempty"`

	// ForceSendFields is a list of field names (e.g. "PageNumber") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "PageNumber") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ImageAnnotationContext) MarshalJSON() ([]byte, error) {
	type NoMethod ImageAnnotationContext
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ImageProperties: Stores image properties, such as dominant colors.
type ImageProperties struct {
	// DominantColors: If present, dominant colors completed successfully.
	DominantColors *DominantColorsAnnotation `json:"dominantColors,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DominantColors") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "DominantColors") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *ImageProperties) MarshalJSON() ([]byte, error) {
	type NoMethod ImageProperties
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// InputConfig: The desired input location and metadata.
type InputConfig struct {
	// GcsSource: The Google Cloud Storage location to read the input from.
	GcsSource *GcsSource `json:"gcsSource,omitempty"`

	// MimeType: The type of the file. Currently only "application/pdf" and
	// "image/tiff"
	// are supported. Wildcards are not supported.
	MimeType string `json:"mimeType,omitempty"`

	// ForceSendFields is a list of field names (e.g. "GcsSource") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "GcsSource") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *InputConfig) MarshalJSON() ([]byte, error) {
	type NoMethod InputConfig
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Landmark: A face-specific landmark (for example, a face feature).
type Landmark struct {
	// Position: Face landmark position.
	Position *Position `json:"position,omitempty"`

	// Type: Face landmark type.
	//
	// Possible values:
	//   "UNKNOWN_LANDMARK" - Unknown face landmark detected. Should not be
	// filled.
	//   "LEFT_EYE" - Left eye.
	//   "RIGHT_EYE" - Right eye.
	//   "LEFT_OF_LEFT_EYEBROW" - Left of left eyebrow.
	//   "RIGHT_OF_LEFT_EYEBROW" - Right of left eyebrow.
	//   "LEFT_OF_RIGHT_EYEBROW" - Left of right eyebrow.
	//   "RIGHT_OF_RIGHT_EYEBROW" - Right of right eyebrow.
	//   "MIDPOINT_BETWEEN_EYES" - Midpoint between eyes.
	//   "NOSE_TIP" - Nose tip.
	//   "UPPER_LIP" - Upper lip.
	//   "LOWER_LIP" - Lower lip.
	//   "MOUTH_LEFT" - Mouth left.
	//   "MOUTH_RIGHT" - Mouth right.
	//   "MOUTH_CENTER" - Mouth center.
	//   "NOSE_BOTTOM_RIGHT" - Nose, bottom right.
	//   "NOSE_BOTTOM_LEFT" - Nose, bottom left.
	//   "NOSE_BOTTOM_CENTER" - Nose, bottom center.
	//   "LEFT_EYE_TOP_BOUNDARY" - Left eye, top boundary.
	//   "LEFT_EYE_RIGHT_CORNER" - Left eye, right corner.
	//   "LEFT_EYE_BOTTOM_BOUNDARY" - Left eye, bottom boundary.
	//   "LEFT_EYE_LEFT_CORNER" - Left eye, left corner.
	//   "RIGHT_EYE_TOP_BOUNDARY" - Right eye, top boundary.
	//   "RIGHT_EYE_RIGHT_CORNER" - Right eye, right corner.
	//   "RIGHT_EYE_BOTTOM_BOUNDARY" - Right eye, bottom boundary.
	//   "RIGHT_EYE_LEFT_CORNER" - Right eye, left corner.
	//   "LEFT_EYEBROW_UPPER_MIDPOINT" - Left eyebrow, upper midpoint.
	//   "RIGHT_EYEBROW_UPPER_MIDPOINT" - Right eyebrow, upper midpoint.
	//   "LEFT_EAR_TRAGION" - Left ear tragion.
	//   "RIGHT_EAR_TRAGION" - Right ear tragion.
	//   "LEFT_EYE_PUPIL" - Left eye pupil.
	//   "RIGHT_EYE_PUPIL" - Right eye pupil.
	//   "FOREHEAD_GLABELLA" - Forehead glabella.
	//   "CHIN_GNATHION" - Chin gnathion.
	//   "CHIN_LEFT_GONION" - Chin left gonion.
	//   "CHIN_RIGHT_GONION" - Chin right gonion.
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Position") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Position") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Landmark) MarshalJSON() ([]byte, error) {
	type NoMethod Landmark
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// LatLng: An object representing a latitude/longitude pair. This is
// expressed as a pair
// of doubles representing degrees latitude and degrees longitude.
// Unless
// specified otherwise, this must conform to the
// <a
// href="http://www.unoosa.org/pdf/icg/2012/template/WGS_84.pdf">WGS84
// st
// andard</a>. Values must be within normalized ranges.
type LatLng struct {
	// Latitude: The latitude in degrees. It must be in the range [-90.0,
	// +90.0].
	Latitude float64 `json:"latitude,omitempty"`

	// Longitude: The longitude in degrees. It must be in the range [-180.0,
	// +180.0].
	Longitude float64 `json:"longitude,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Latitude") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Latitude") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *LatLng) MarshalJSON() ([]byte, error) {
	type NoMethod LatLng
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *LatLng) UnmarshalJSON(data []byte) error {
	type NoMethod LatLng
	var s1 struct {
		Latitude  gensupport.JSONFloat64 `json:"latitude"`
		Longitude gensupport.JSONFloat64 `json:"longitude"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Latitude = float64(s1.Latitude)
	s.Longitude = float64(s1.Longitude)
	return nil
}

// LocationInfo: Detected entity location information.
type LocationInfo struct {
	// LatLng: lat/long location coordinates.
	LatLng *LatLng `json:"latLng,omitempty"`

	// ForceSendFields is a list of field names (e.g. "LatLng") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "LatLng") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *LocationInfo) MarshalJSON() ([]byte, error) {
	type NoMethod LocationInfo
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// NormalizedVertex: A vertex represents a 2D point in the image.
// NOTE: the normalized vertex coordinates are relative to the original
// image
// and range from 0 to 1.
type NormalizedVertex struct {
	// X: X coordinate.
	X float64 `json:"x,omitempty"`

	// Y: Y coordinate.
	Y float64 `json:"y,omitempty"`

	// ForceSendFields is a list of field names (e.g. "X") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "X") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *NormalizedVertex) MarshalJSON() ([]byte, error) {
	type NoMethod NormalizedVertex
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *NormalizedVertex) UnmarshalJSON(data []byte) error {
	type NoMethod NormalizedVertex
	var s1 struct {
		X gensupport.JSONFloat64 `json:"x"`
		Y gensupport.JSONFloat64 `json:"y"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.X = float64(s1.X)
	s.Y = float64(s1.Y)
	return nil
}

// Operation: This resource represents a long-running operation that is
// the result of a
// network API call.
type Operation struct {
	// Done: If the value is `false`, it means the operation is still in
	// progress.
	// If `true`, the operation is completed, and either `error` or
	// `response` is
	// available.
	Done bool `json:"done,omitempty"`

	// Error: The error result of the operation in case of failure or
	// cancellation.
	Error *Status `json:"error,omitempty"`

	// Metadata: Service-specific metadata associated with the operation.
	// It typically
	// contains progress information and common metadata such as create
	// time.
	// Some services might not provide such metadata.  Any method that
	// returns a
	// long-running operation should document the metadata type, if any.
	Metadata googleapi.RawMessage `json:"metadata,omitempty"`

	// Name: The server-assigned name, which is only unique within the same
	// service that
	// originally returns it. If you use the default HTTP mapping,
	// the
	// `name` should have the format of `operations/some/unique/name`.
	Name string `json:"name,omitempty"`

	// Response: The normal response of the operation in case of success.
	// If the original
	// method returns no data on success, such as `Delete`, the response
	// is
	// `google.protobuf.Empty`.  If the original method is
	// standard
	// `Get`/`Create`/`Update`, the response should be the resource.  For
	// other
	// methods, the response should have the type `XxxResponse`, where
	// `Xxx`
	// is the original method name.  For example, if the original method
	// name
	// is `TakeSnapshot()`, the inferred response type
	// is
	// `TakeSnapshotResponse`.
	Response googleapi.RawMessage `json:"response,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Done") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Done") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Operation) MarshalJSON() ([]byte, error) {
	type NoMethod Operation
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// OperationMetadata: Contains metadata for the BatchAnnotateImages
// operation.
type OperationMetadata struct {
	// CreateTime: The time when the batch request was received.
	CreateTime string `json:"createTime,omitempty"`

	// State: Current state of the batch operation.
	//
	// Possible values:
	//   "STATE_UNSPECIFIED" - Invalid.
	//   "CREATED" - Request is received.
	//   "RUNNING" - Request is actively being processed.
	//   "DONE" - The batch processing is done.
	//   "CANCELLED" - The batch processing was cancelled.
	State string `json:"state,omitempty"`

	// UpdateTime: The time when the operation result was last updated.
	UpdateTime string `json:"updateTime,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CreateTime") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "CreateTime") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *OperationMetadata) MarshalJSON() ([]byte, error) {
	type NoMethod OperationMetadata
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// OutputConfig: The desired output location and metadata.
type OutputConfig struct {
	// BatchSize: The max number of response protos to put into each output
	// JSON file on
	// Google Cloud Storage.
	// The valid range is [1, 100]. If not specified, the default value is
	// 20.
	//
	// For example, for one pdf file with 100 pages, 100 response protos
	// will
	// be generated. If `batch_size` = 20, then 5 json files each
	// containing 20 response protos will be written under the
	// prefix
	// `gcs_destination`.`uri`.
	//
	// Currently, batch_size only applies to GcsDestination, with potential
	// future
	// support for other output configurations.
	BatchSize int64 `json:"batchSize,omitempty"`

	// GcsDestination: The Google Cloud Storage location to write the
	// output(s) to.
	GcsDestination *GcsDestination `json:"gcsDestination,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BatchSize") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BatchSize") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *OutputConfig) MarshalJSON() ([]byte, error) {
	type NoMethod OutputConfig
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Page: Detected page from OCR.
type Page struct {
	// Blocks: List of blocks of text, images etc on this page.
	Blocks []*Block `json:"blocks,omitempty"`

	// Confidence: Confidence of the OCR results on the page. Range [0, 1].
	Confidence float64 `json:"confidence,omitempty"`

	// Height: Page height. For PDFs the unit is points. For images
	// (including
	// TIFFs) the unit is pixels.
	Height int64 `json:"height,omitempty"`

	// Property: Additional information detected on the page.
	Property *TextProperty `json:"property,omitempty"`

	// Width: Page width. For PDFs the unit is points. For images
	// (including
	// TIFFs) the unit is pixels.
	Width int64 `json:"width,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Blocks") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Blocks") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Page) MarshalJSON() ([]byte, error) {
	type NoMethod Page
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *Page) UnmarshalJSON(data []byte) error {
	type NoMethod Page
	var s1 struct {
		Confidence gensupport.JSONFloat64 `json:"confidence"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	return nil
}

// Paragraph: Structural unit of text representing a number of words in
// certain order.
type Paragraph struct {
	// BoundingBox: The bounding box for the paragraph.
	// The vertices are in the order of top-left, top-right,
	// bottom-right,
	// bottom-left. When a rotation of the bounding box is detected the
	// rotation
	// is represented as around the top-left corner as defined when the text
	// is
	// read in the 'natural' orientation.
	// For example:
	//   * when the text is horizontal it might look like:
	//      0----1
	//      |    |
	//      3----2
	//   * when it's rotated 180 degrees around the top-left corner it
	// becomes:
	//      2----3
	//      |    |
	//      1----0
	//   and the vertice order will still be (0, 1, 2, 3).
	BoundingBox *BoundingPoly `json:"boundingBox,omitempty"`

	// Confidence: Confidence of the OCR results for the paragraph. Range
	// [0, 1].
	Confidence float64 `json:"confidence,omitempty"`

	// Property: Additional information detected for the paragraph.
	Property *TextProperty `json:"property,omitempty"`

	// Words: List of words in this paragraph.
	Words []*Word `json:"words,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BoundingBox") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BoundingBox") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Paragraph) MarshalJSON() ([]byte, error) {
	type NoMethod Paragraph
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *Paragraph) UnmarshalJSON(data []byte) error {
	type NoMethod Paragraph
	var s1 struct {
		Confidence gensupport.JSONFloat64 `json:"confidence"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	return nil
}

// Position: A 3D position in the image, used primarily for Face
// detection landmarks.
// A valid Position must have both x and y coordinates.
// The position coordinates are in the same scale as the original image.
type Position struct {
	// X: X coordinate.
	X float64 `json:"x,omitempty"`

	// Y: Y coordinate.
	Y float64 `json:"y,omitempty"`

	// Z: Z coordinate (or depth).
	Z float64 `json:"z,omitempty"`

	// ForceSendFields is a list of field names (e.g. "X") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "X") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Position) MarshalJSON() ([]byte, error) {
	type NoMethod Position
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *Position) UnmarshalJSON(data []byte) error {
	type NoMethod Position
	var s1 struct {
		X gensupport.JSONFloat64 `json:"x"`
		Y gensupport.JSONFloat64 `json:"y"`
		Z gensupport.JSONFloat64 `json:"z"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.X = float64(s1.X)
	s.Y = float64(s1.Y)
	s.Z = float64(s1.Z)
	return nil
}

// Property: A `Property` consists of a user-supplied name/value pair.
type Property struct {
	// Name: Name of the property.
	Name string `json:"name,omitempty"`

	// Uint64Value: Value of numeric properties.
	Uint64Value uint64 `json:"uint64Value,omitempty,string"`

	// Value: Value of the property.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Name") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Name") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Property) MarshalJSON() ([]byte, error) {
	type NoMethod Property
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// SafeSearchAnnotation: Set of features pertaining to the image,
// computed by computer vision
// methods over safe-search verticals (for example, adult, spoof,
// medical,
// violence).
type SafeSearchAnnotation struct {
	// Adult: Represents the adult content likelihood for the image. Adult
	// content may
	// contain elements such as nudity, pornographic images or cartoons,
	// or
	// sexual activities.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	Adult string `json:"adult,omitempty"`

	// Medical: Likelihood that this is a medical image.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	Medical string `json:"medical,omitempty"`

	// Racy: Likelihood that the request image contains racy content. Racy
	// content may
	// include (but is not limited to) skimpy or sheer clothing,
	// strategically
	// covered nudity, lewd or provocative poses, or close-ups of
	// sensitive
	// body areas.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	Racy string `json:"racy,omitempty"`

	// Spoof: Spoof likelihood. The likelihood that an modification
	// was made to the image's canonical version to make it appear
	// funny or offensive.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	Spoof string `json:"spoof,omitempty"`

	// Violence: Likelihood that this image contains violent content.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - It is very unlikely that the image belongs to the
	// specified vertical.
	//   "UNLIKELY" - It is unlikely that the image belongs to the specified
	// vertical.
	//   "POSSIBLE" - It is possible that the image belongs to the specified
	// vertical.
	//   "LIKELY" - It is likely that the image belongs to the specified
	// vertical.
	//   "VERY_LIKELY" - It is very likely that the image belongs to the
	// specified vertical.
	Violence string `json:"violence,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Adult") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Adult") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *SafeSearchAnnotation) MarshalJSON() ([]byte, error) {
	type NoMethod SafeSearchAnnotation
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Status: The `Status` type defines a logical error model that is
// suitable for different
// programming environments, including REST APIs and RPC APIs. It is
// used by
// [gRPC](https://github.com/grpc). The error model is designed to
// be:
//
// - Simple to use and understand for most users
// - Flexible enough to meet unexpected needs
//
// # Overview
//
// The `Status` message contains three pieces of data: error code, error
// message,
// and error details. The error code should be an enum value
// of
// google.rpc.Code, but it may accept additional error codes if needed.
// The
// error message should be a developer-facing English message that
// helps
// developers *understand* and *resolve* the error. If a localized
// user-facing
// error message is needed, put the localized message in the error
// details or
// localize it in the client. The optional error details may contain
// arbitrary
// information about the error. There is a predefined set of error
// detail types
// in the package `google.rpc` that can be used for common error
// conditions.
//
// # Language mapping
//
// The `Status` message is the logical representation of the error
// model, but it
// is not necessarily the actual wire format. When the `Status` message
// is
// exposed in different client libraries and different wire protocols,
// it can be
// mapped differently. For example, it will likely be mapped to some
// exceptions
// in Java, but more likely mapped to some error codes in C.
//
// # Other uses
//
// The error model and the `Status` message can be used in a variety
// of
// environments, either with or without APIs, to provide a
// consistent developer experience across different
// environments.
//
// Example uses of this error model include:
//
// - Partial errors. If a service needs to return partial errors to the
// client,
//     it may embed the `Status` in the normal response to indicate the
// partial
//     errors.
//
// - Workflow errors. A typical workflow has multiple steps. Each step
// may
//     have a `Status` message for error reporting.
//
// - Batch operations. If a client uses batch request and batch
// response, the
//     `Status` message should be used directly inside batch response,
// one for
//     each error sub-response.
//
// - Asynchronous operations. If an API call embeds asynchronous
// operation
//     results in its response, the status of those operations should
// be
//     represented directly using the `Status` message.
//
// - Logging. If some API errors are stored in logs, the message
// `Status` could
//     be used directly after any stripping needed for security/privacy
// reasons.
type Status struct {
	// Code: The status code, which should be an enum value of
	// google.rpc.Code.
	Code int64 `json:"code,omitempty"`

	// Details: A list of messages that carry the error details.  There is a
	// common set of
	// message types for APIs to use.
	Details []googleapi.RawMessage `json:"details,omitempty"`

	// Message: A developer-facing error message, which should be in
	// English. Any
	// user-facing error message should be localized and sent in
	// the
	// google.rpc.Status.details field, or localized by the client.
	Message string `json:"message,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Code") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Code") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Status) MarshalJSON() ([]byte, error) {
	type NoMethod Status
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Symbol: A single symbol representation.
type Symbol struct {
	// BoundingBox: The bounding box for the symbol.
	// The vertices are in the order of top-left, top-right,
	// bottom-right,
	// bottom-left. When a rotation of the bounding box is detected the
	// rotation
	// is represented as around the top-left corner as defined when the text
	// is
	// read in the 'natural' orientation.
	// For example:
	//   * when the text is horizontal it might look like:
	//      0----1
	//      |    |
	//      3----2
	//   * when it's rotated 180 degrees around the top-left corner it
	// becomes:
	//      2----3
	//      |    |
	//      1----0
	//   and the vertice order will still be (0, 1, 2, 3).
	BoundingBox *BoundingPoly `json:"boundingBox,omitempty"`

	// Confidence: Confidence of the OCR results for the symbol. Range [0,
	// 1].
	Confidence float64 `json:"confidence,omitempty"`

	// Property: Additional information detected for the symbol.
	Property *TextProperty `json:"property,omitempty"`

	// Text: The actual UTF-8 representation of the symbol.
	Text string `json:"text,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BoundingBox") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BoundingBox") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Symbol) MarshalJSON() ([]byte, error) {
	type NoMethod Symbol
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *Symbol) UnmarshalJSON(data []byte) error {
	type NoMethod Symbol
	var s1 struct {
		Confidence gensupport.JSONFloat64 `json:"confidence"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	return nil
}

// TextAnnotation: TextAnnotation contains a structured representation
// of OCR extracted text.
// The hierarchy of an OCR extracted text structure is like this:
//     TextAnnotation -> Page -> Block -> Paragraph -> Word ->
// Symbol
// Each structural component, starting from Page, may further have their
// own
// properties. Properties describe detected languages, breaks etc..
// Please refer
// to the TextAnnotation.TextProperty message definition below for
// more
// detail.
type TextAnnotation struct {
	// Pages: List of pages detected by OCR.
	Pages []*Page `json:"pages,omitempty"`

	// Text: UTF-8 text detected on the pages.
	Text string `json:"text,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Pages") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Pages") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *TextAnnotation) MarshalJSON() ([]byte, error) {
	type NoMethod TextAnnotation
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// TextProperty: Additional information detected on the structural
// component.
type TextProperty struct {
	// DetectedBreak: Detected start or end of a text segment.
	DetectedBreak *DetectedBreak `json:"detectedBreak,omitempty"`

	// DetectedLanguages: A list of detected languages together with
	// confidence.
	DetectedLanguages []*DetectedLanguage `json:"detectedLanguages,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DetectedBreak") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "DetectedBreak") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *TextProperty) MarshalJSON() ([]byte, error) {
	type NoMethod TextProperty
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Vertex: A vertex represents a 2D point in the image.
// NOTE: the vertex coordinates are in the same scale as the original
// image.
type Vertex struct {
	// X: X coordinate.
	X int64 `json:"x,omitempty"`

	// Y: Y coordinate.
	Y int64 `json:"y,omitempty"`

	// ForceSendFields is a list of field names (e.g. "X") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "X") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Vertex) MarshalJSON() ([]byte, error) {
	type NoMethod Vertex
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// WebDetection: Relevant information for the image from the Internet.
type WebDetection struct {
	// BestGuessLabels: The service's best guess as to the topic of the
	// request image.
	// Inferred from similar images on the open web.
	BestGuessLabels []*WebLabel `json:"bestGuessLabels,omitempty"`

	// FullMatchingImages: Fully matching images from the Internet.
	// Can include resized copies of the query image.
	FullMatchingImages []*WebImage `json:"fullMatchingImages,omitempty"`

	// PagesWithMatchingImages: Web pages containing the matching images
	// from the Internet.
	PagesWithMatchingImages []*WebPage `json:"pagesWithMatchingImages,omitempty"`

	// PartialMatchingImages: Partial matching images from the
	// Internet.
	// Those images are similar enough to share some key-point features.
	// For
	// example an original image will likely have partial matching for its
	// crops.
	PartialMatchingImages []*WebImage `json:"partialMatchingImages,omitempty"`

	// VisuallySimilarImages: The visually similar image results.
	VisuallySimilarImages []*WebImage `json:"visuallySimilarImages,omitempty"`

	// WebEntities: Deduced entities from similar images on the Internet.
	WebEntities []*WebEntity `json:"webEntities,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BestGuessLabels") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BestGuessLabels") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *WebDetection) MarshalJSON() ([]byte, error) {
	type NoMethod WebDetection
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// WebEntity: Entity deduced from similar images on the Internet.
type WebEntity struct {
	// Description: Canonical description of the entity, in English.
	Description string `json:"description,omitempty"`

	// EntityId: Opaque entity ID.
	EntityId string `json:"entityId,omitempty"`

	// Score: Overall relevancy score for the entity.
	// Not normalized and not comparable across different image queries.
	Score float64 `json:"score,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Description") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Description") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *WebEntity) MarshalJSON() ([]byte, error) {
	type NoMethod WebEntity
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *WebEntity) UnmarshalJSON(data []byte) error {
	type NoMethod WebEntity
	var s1 struct {
		Score gensupport.JSONFloat64 `json:"score"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Score = float64(s1.Score)
	return nil
}

// WebImage: Metadata for online images.
type WebImage struct {
	// Score: (Deprecated) Overall relevancy score for the image.
	Score float64 `json:"score,omitempty"`

	// Url: The result image URL.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Score") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Score") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *WebImage) MarshalJSON() ([]byte, error) {
	type NoMethod WebImage
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *WebImage) UnmarshalJSON(data []byte) error {
	type NoMethod WebImage
	var s1 struct {
		Score gensupport.JSONFloat64 `json:"score"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Score = float64(s1.Score)
	return nil
}

// WebLabel: Label to provide extra metadata for the web detection.
type WebLabel struct {
	// Label: Label for extra metadata.
	Label string `json:"label,omitempty"`

	// LanguageCode: The BCP-47 language code for `label`, such as "en-US"
	// or "sr-Latn".
	// For more information,
	// see
	// http://www.unicode.org/reports/tr35/#Unicode_locale_identifier.
	LanguageCode string `json:"languageCode,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Label") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Label") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *WebLabel) MarshalJSON() ([]byte, error) {
	type NoMethod WebLabel
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// WebPage: Metadata for web pages.
type WebPage struct {
	// FullMatchingImages: Fully matching images on the page.
	// Can include resized copies of the query image.
	FullMatchingImages []*WebImage `json:"fullMatchingImages,omitempty"`

	// PageTitle: Title for the web page, may contain HTML markups.
	PageTitle string `json:"pageTitle,omitempty"`

	// PartialMatchingImages: Partial matching images on the page.
	// Those images are similar enough to share some key-point features.
	// For
	// example an original image will likely have partial matching for
	// its
	// crops.
	PartialMatchingImages []*WebImage `json:"partialMatchingImages,omitempty"`

	// Score: (Deprecated) Overall relevancy score for the web page.
	Score float64 `json:"score,omitempty"`

	// Url: The result web page URL.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FullMatchingImages")
	// to unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "FullMatchingImages") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *WebPage) MarshalJSON() ([]byte, error) {
	type NoMethod WebPage
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *WebPage) UnmarshalJSON(data []byte) error {
	type NoMethod WebPage
	var s1 struct {
		Score gensupport.JSONFloat64 `json:"score"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Score = float64(s1.Score)
	return nil
}

// Word: A word representation.
type Word struct {
	// BoundingBox: The bounding box for the word.
	// The vertices are in the order of top-left, top-right,
	// bottom-right,
	// bottom-left. When a rotation of the bounding box is detected the
	// rotation
	// is represented as around the top-left corner as defined when the text
	// is
	// read in the 'natural' orientation.
	// For example:
	//   * when the text is horizontal it might look like:
	//      0----1
	//      |    |
	//      3----2
	//   * when it's rotated 180 degrees around the top-left corner it
	// becomes:
	//      2----3
	//      |    |
	//      1----0
	//   and the vertice order will still be (0, 1, 2, 3).
	BoundingBox *BoundingPoly `json:"boundingBox,omitempty"`

	// Confidence: Confidence of the OCR results for the word. Range [0, 1].
	Confidence float64 `json:"confidence,omitempty"`

	// Property: Additional information detected for the word.
	Property *TextProperty `json:"property,omitempty"`

	// Symbols: List of symbols in the word.
	// The order of the symbols follows the natural reading order.
	Symbols []*Symbol `json:"symbols,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BoundingBox") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BoundingBox") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Word) MarshalJSON() ([]byte, error) {
	type NoMethod Word
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *Word) UnmarshalJSON(data []byte) error {
	type NoMethod Word
	var s1 struct {
		Confidence gensupport.JSONFloat64 `json:"confidence"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Confidence = float64(s1.Confidence)
	return nil
}

// method id "vision.files.asyncBatchAnnotate":

type FilesAsyncBatchAnnotateCall struct {
	s                                                        *Service
	googlecloudvisionv1p2beta1asyncbatchannotatefilesrequest *GoogleCloudVisionV1p2beta1AsyncBatchAnnotateFilesRequest
	urlParams_                                               gensupport.URLParams
	ctx_                                                     context.Context
	header_                                                  http.Header
}

// AsyncBatchAnnotate: Run asynchronous image detection and annotation
// for a list of generic
// files, such as PDF files, which may contain multiple pages and
// multiple
// images per page. Progress and results can be retrieved through
// the
// `google.longrunning.Operations` interface.
// `Operation.metadata` contains `OperationMetadata`
// (metadata).
// `Operation.response` contains `AsyncBatchAnnotateFilesResponse`
// (results).
func (r *FilesService) AsyncBatchAnnotate(googlecloudvisionv1p2beta1asyncbatchannotatefilesrequest *GoogleCloudVisionV1p2beta1AsyncBatchAnnotateFilesRequest) *FilesAsyncBatchAnnotateCall {
	c := &FilesAsyncBatchAnnotateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.googlecloudvisionv1p2beta1asyncbatchannotatefilesrequest = googlecloudvisionv1p2beta1asyncbatchannotatefilesrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *FilesAsyncBatchAnnotateCall) Fields(s ...googleapi.Field) *FilesAsyncBatchAnnotateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *FilesAsyncBatchAnnotateCall) Context(ctx context.Context) *FilesAsyncBatchAnnotateCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *FilesAsyncBatchAnnotateCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *FilesAsyncBatchAnnotateCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.googlecloudvisionv1p2beta1asyncbatchannotatefilesrequest)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1p2beta1/files:asyncBatchAnnotate")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "vision.files.asyncBatchAnnotate" call.
// Exactly one of *Operation or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Operation.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *FilesAsyncBatchAnnotateCall) Do(opts ...googleapi.CallOption) (*Operation, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Operation{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Run asynchronous image detection and annotation for a list of generic\nfiles, such as PDF files, which may contain multiple pages and multiple\nimages per page. Progress and results can be retrieved through the\n`google.longrunning.Operations` interface.\n`Operation.metadata` contains `OperationMetadata` (metadata).\n`Operation.response` contains `AsyncBatchAnnotateFilesResponse` (results).",
	//   "flatPath": "v1p2beta1/files:asyncBatchAnnotate",
	//   "httpMethod": "POST",
	//   "id": "vision.files.asyncBatchAnnotate",
	//   "parameterOrder": [],
	//   "parameters": {},
	//   "path": "v1p2beta1/files:asyncBatchAnnotate",
	//   "request": {
	//     "$ref": "GoogleCloudVisionV1p2beta1AsyncBatchAnnotateFilesRequest"
	//   },
	//   "response": {
	//     "$ref": "Operation"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/cloud-vision"
	//   ]
	// }

}

// method id "vision.images.annotate":

type ImagesAnnotateCall struct {
	s                                                    *Service
	googlecloudvisionv1p2beta1batchannotateimagesrequest *GoogleCloudVisionV1p2beta1BatchAnnotateImagesRequest
	urlParams_                                           gensupport.URLParams
	ctx_                                                 context.Context
	header_                                              http.Header
}

// Annotate: Run image detection and annotation for a batch of images.
func (r *ImagesService) Annotate(googlecloudvisionv1p2beta1batchannotateimagesrequest *GoogleCloudVisionV1p2beta1BatchAnnotateImagesRequest) *ImagesAnnotateCall {
	c := &ImagesAnnotateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.googlecloudvisionv1p2beta1batchannotateimagesrequest = googlecloudvisionv1p2beta1batchannotateimagesrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ImagesAnnotateCall) Fields(s ...googleapi.Field) *ImagesAnnotateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ImagesAnnotateCall) Context(ctx context.Context) *ImagesAnnotateCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ImagesAnnotateCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ImagesAnnotateCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.googlecloudvisionv1p2beta1batchannotateimagesrequest)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1p2beta1/images:annotate")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "vision.images.annotate" call.
// Exactly one of *GoogleCloudVisionV1p2beta1BatchAnnotateImagesResponse
// or error will be non-nil. Any non-2xx status code is an error.
// Response headers are in either
// *GoogleCloudVisionV1p2beta1BatchAnnotateImagesResponse.ServerResponse.
// Header or (if a response was returned at all) in
// error.(*googleapi.Error).Header. Use googleapi.IsNotModified to check
// whether the returned error was because http.StatusNotModified was
// returned.
func (c *ImagesAnnotateCall) Do(opts ...googleapi.CallOption) (*GoogleCloudVisionV1p2beta1BatchAnnotateImagesResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &GoogleCloudVisionV1p2beta1BatchAnnotateImagesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Run image detection and annotation for a batch of images.",
	//   "flatPath": "v1p2beta1/images:annotate",
	//   "httpMethod": "POST",
	//   "id": "vision.images.annotate",
	//   "parameterOrder": [],
	//   "parameters": {},
	//   "path": "v1p2beta1/images:annotate",
	//   "request": {
	//     "$ref": "GoogleCloudVisionV1p2beta1BatchAnnotateImagesRequest"
	//   },
	//   "response": {
	//     "$ref": "GoogleCloudVisionV1p2beta1BatchAnnotateImagesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/cloud-vision"
	//   ]
	// }

}
