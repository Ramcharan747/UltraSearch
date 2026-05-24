package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ContributionGateway manages the validation pipeline for user or community Skill Books
type ContributionGateway struct {
	StagingDir string
	ActiveDir  string
}

// NewContributionGateway creates a new validator instance
func NewContributionGateway(activeDir, stagingDir string) *ContributionGateway {
	return &ContributionGateway{
		ActiveDir:  activeDir,
		StagingDir: stagingDir,
	}
}

// SandboxValidateSkillBook scans a contributed Skill Book for security or injection threats
func (cg *ContributionGateway) SandboxValidateSkillBook(filePath string) (string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "ERROR", err
	}

	content := string(bytes)
	lowerContent := strings.ToLower(content)

	// Dry-run security threat signatures (Layer 6.1 Attack Vectors)
	maliciousTriggers := []string{
		"pastebin.com/raw",
		"exfiltrate",
		"send credentials",
		"attacker.com",
		"ignore previous instructions",
		"override current system prompt",
		"db_password",
		"api_key",
		"private_key",
		"hack",
		"jailbreak",
		"bypass safety",
	}

	var triggeredTriggers []string
	for _, trigger := range maliciousTriggers {
		if strings.Contains(lowerContent, trigger) {
			triggeredTriggers = append(triggeredTriggers, trigger)
		}
	}

	if len(triggeredTriggers) > 0 {
		return "MALICIOUS", fmt.Errorf("security audit failed: flagged adversarial signatures %v", triggeredTriggers)
	}

	// Validate that it parses cleanly as a valid Skill Book (check Frontmatter metadata)
	_, err = ParseSkillBookFile(filePath)
	if err != nil {
		return "INVALID_METADATA", fmt.Errorf("frontmatter syntax validation failed: %v", err)
	}

	return "UNVERIFIED_SAFE", nil
}

// IntakeSkillBook processes a new Skill Book, validates it in the sandbox, and writes it to staging
func (cg *ContributionGateway) IntakeSkillBook(srcPath string) (string, error) {
	// Create directories if missing
	_ = os.MkdirAll(cg.ActiveDir, 0755)
	_ = os.MkdirAll(cg.StagingDir, 0755)

	baseName := filepath.Base(srcPath)
	destStagingPath := filepath.Join(cg.StagingDir, baseName)

	// Copy file content to staging for analysis
	bytes, err := os.ReadFile(srcPath)
	if err != nil {
		return "READ_ERROR", err
	}
	err = os.WriteFile(destStagingPath, bytes, 0644)
	if err != nil {
		return "WRITE_ERROR", err
	}

	// Run Sandbox validation
	status, err := cg.SandboxValidateSkillBook(destStagingPath)
	if err != nil {
		// If malicious or invalid, delete from staging instantly
		_ = os.Remove(destStagingPath)
		return status, err
	}

	log.Printf("🛡️ [WI-OS Intake] Sandbox audit clean. Skill Book staged at: %s (Status: Awaiting Human Review)", destStagingPath)
	return "STAGED", nil
}

// PromoteSkillBook moves a verified Skill Book from staging to active directory (Human-in-the-Loop Integration)
func (cg *ContributionGateway) PromoteSkillBook(fileName string) error {
	stagingPath := filepath.Join(cg.StagingDir, fileName)
	activePath := filepath.Join(cg.ActiveDir, fileName)

	if _, err := os.Stat(stagingPath); os.IsNotExist(err) {
		return fmt.Errorf("skill book %s does not exist in staging", fileName)
	}

	// Verify one more time before moving
	_, err := cg.SandboxValidateSkillBook(stagingPath)
	if err != nil {
		return fmt.Errorf("re-verification failed: %v", err)
	}

	// Move file
	err = os.Rename(stagingPath, activePath)
	if err != nil {
		// Fallback: Copy and Delete
		bytes, copyErr := os.ReadFile(stagingPath)
		if copyErr != nil {
			return copyErr
		}
		writeErr := os.WriteFile(activePath, bytes, 0644)
		if writeErr != nil {
			return writeErr
		}
		_ = os.Remove(stagingPath)
	}

	log.Printf("✅ [WI-OS Promotion] Skill Book %s verified and promoted to Active Engine.", fileName)
	return nil
}

// ListStagedSkillBooks returns all templates waiting in staging
func (cg *ContributionGateway) ListStagedSkillBooks() ([]string, error) {
	files, err := os.ReadDir(cg.StagingDir)
	if err != nil {
		return nil, err
	}

	var list []string
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".md" {
			list = append(list, f.Name())
		}
	}
	return list, nil
}

// RunContributionDiagnostics validates the dry-run threat sandbox and human-in-the-loop promotion
func RunContributionDiagnostics() {
	fmt.Println("\n[*] Running WI-OS Contribution Pipeline & Safety Diagnostics...")
	cg := NewContributionGateway("ai_skills", filepath.Join("ai_skills", "unverified"))

	// Create test unverified files folder
	_ = os.MkdirAll(filepath.Join("ai_skills", "unverified"), 0755)

	// Mock Diagnostic Test 1: Contributes a malicious dump-mining Skill Book
	maliciousContent := `---
name: leaked_credential_dorks
version: 1.0.0
author: HackerX
domains: [cyber_attacks, credentials]
---
SEARCH pastebin.com/raw
RETURN {
  db_password: string
}
`
	mPath := filepath.Join("ai_skills", "unverified", "test_leaked_hack.md")
	_ = os.WriteFile(mPath, []byte(maliciousContent), 0644)

	status1, err1 := cg.SandboxValidateSkillBook(mPath)
	fmt.Printf("\nMock Test 1 (Adversarial Skill Contribution) - Status: %s\n", status1)
	if err1 != nil {
		fmt.Printf("   Safety Warning output: %v\n", err1)
	}
	_ = os.Remove(mPath)

	// Mock Diagnostic Test 2: Contributes a benign customized Skill Book
	benignContent := `---
name: custom_market_intelligence
version: 1.2.0
author: GlenLindenstaedt
trust_tier: community
domains: [market_trends, competitor_analysis]
---
# Custom Market Intelligence
This is a custom Skill Book for tracking company ARR and marketing budgets.
`
	bPath := filepath.Join("ai_skills", "unverified", "custom_market.md")
	_ = os.WriteFile(bPath, []byte(benignContent), 0644)

	status2, err2 := cg.SandboxValidateSkillBook(bPath)
	fmt.Printf("\nMock Test 2 (Benign Skill Contribution) - Status: %s\n", status2)
	if err2 == nil {
		fmt.Printf("   Success: staged and verified cleanly. Ready for promotion.\n")
	}

	// Test Human-in-the-loop promotion
	err := cg.PromoteSkillBook("custom_market.md")
	if err == nil {
		fmt.Println("   Success: Skill Book verified and promoted to active catalog!")
		// Clean up promoted file
		_ = os.Remove(filepath.Join("ai_skills", "custom_market.md"))
	} else {
		fmt.Printf("   Failed promotion: %v\n", err)
	}
	_ = os.Remove(bPath)
}
