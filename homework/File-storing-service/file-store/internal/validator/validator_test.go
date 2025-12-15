package validator

import (
	"testing"
)

func TestValidateUploadRequest_Success(t *testing.T) {
	err := ValidateUploadRequest("test.txt", "student1", "assignment1", 100, 1000)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateUploadRequest_EmptyFilename(t *testing.T) {
	err := ValidateUploadRequest("", "student1", "assignment1", 100, 1000)
	if err == nil {
		t.Error("Expected error for empty filename")
	}
}

func TestValidateUploadRequest_EmptyStudentID(t *testing.T) {
	err := ValidateUploadRequest("test.txt", "", "assignment1", 100, 1000)
	if err == nil {
		t.Error("Expected error for empty student_id")
	}
}

func TestValidateUploadRequest_EmptyAssignmentID(t *testing.T) {
	err := ValidateUploadRequest("test.txt", "student1", "", 100, 1000)
	if err == nil {
		t.Error("Expected error for empty assignment_id")
	}
}

func TestValidateUploadRequest_FileTooLarge(t *testing.T) {
	err := ValidateUploadRequest("test.txt", "student1", "assignment1", 2000, 1000)
	if err == nil {
		t.Error("Expected error for file too large")
	}
}

func TestValidateMimeType_Allowed(t *testing.T) {
	err := ValidateMimeType("text/plain")
	if err != nil {
		t.Errorf("Expected no error for text/plain, got: %v", err)
	}

	err = ValidateMimeType("application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	if err != nil {
		t.Errorf("Expected no error for docx, got: %v", err)
	}
}

func TestValidateMimeType_NotAllowed(t *testing.T) {
	err := ValidateMimeType("application/pdf")
	if err == nil {
		t.Error("Expected error for unsupported MIME type")
	}
}

func TestDetectMimeType_FromExtension(t *testing.T) {
	mime := DetectMimeType("test.txt", "")
	if mime != "text/plain" {
		t.Errorf("Expected text/plain, got %s", mime)
	}

	mime = DetectMimeType("test.docx", "")
	if mime != "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		t.Errorf("Expected docx MIME type, got %s", mime)
	}
}

func TestDetectMimeType_FromProvided(t *testing.T) {
	mime := DetectMimeType("test.txt", "text/plain")
	if mime != "text/plain" {
		t.Errorf("Expected text/plain, got %s", mime)
	}
}

func TestDetectMimeType_Unknown(t *testing.T) {
	mime := DetectMimeType("test.unknown", "")
	if mime != "application/octet-stream" {
		t.Errorf("Expected application/octet-stream, got %s", mime)
	}
}

