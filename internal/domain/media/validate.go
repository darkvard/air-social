package media

import (
	"fmt"
	"slices"
	"strings"

	"air-social/pkg"
)

func IsDomainFeatureValid(params ConfirmParams) bool {
	features, ok := ValidDomainFeatures[params.Domain]
	if !ok {
		return false
	}

	return slices.Contains(features, params.Feature)
}

func validatePresignParams(params PresignParams) (Rule, error) {
	var empty Rule

	features, ok := MediaUploadRules[params.Domain]
	if !ok {
		return empty, fmt.Errorf("domain '%s' is not supported: %w", params.Domain, pkg.ErrBadRequest)
	}

	rule, ok := features[params.Feature]
	if !ok {
		return empty, fmt.Errorf("feature '%s' is not supported for domain '%s': %w", params.Feature, params.Domain, pkg.ErrBadRequest)
	}

	if params.FileSize > rule.MaxBytes {
		return empty, fmt.Errorf("file size %d bytes exceeds limit of %d bytes: %w", params.FileSize, rule.MaxBytes, pkg.ErrFileTooLarge)
	}

	if !slices.Contains(rule.AllowedTypes, params.FileType) {
		return empty, fmt.Errorf("file type '%s' is not allowed (allowed types: %v): %w", params.FileType, rule.AllowedTypes, pkg.ErrBadRequest)
	}

	return rule, nil
}

func validateConfirmParams(input ConfirmParams) error {
	allowedFeatures, ok := ValidDomainFeatures[input.Domain]
	if !ok {
		return fmt.Errorf("domain '%s' not supported: %w", input.Domain, pkg.ErrBadRequest)
	}

	if !slices.Contains(allowedFeatures, input.Feature) {
		return fmt.Errorf("feature '%s' not allowed for domain '%s': %w", input.Feature, input.Domain, pkg.ErrBadRequest)
	}

	expectedPrefix := fmt.Sprintf("%s/%d/%s/", input.Domain, input.EntityID, input.Feature)

	if !strings.HasPrefix(input.ObjectKey, expectedPrefix) {
		return fmt.Errorf("object key mismatch (expected prefix: %s): %w", expectedPrefix, pkg.ErrBadRequest)
	}

	return nil
}
