package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
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

// DownloadCommunitySkill fetches, stages, and security-validates community markdown Skill Books from raw URLs or ZIP repositories
func (cg *ContributionGateway) DownloadCommunitySkill(targetURL string) ([]string, error) {
	_ = os.MkdirAll(cg.StagingDir, 0755)

	// If direct raw markdown file
	if strings.HasSuffix(strings.ToLower(targetURL), ".md") || strings.Contains(targetURL, "/raw/") {
		resp, err := http.Get(targetURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch raw URL: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP error fetching URL: %d %s", resp.StatusCode, resp.Status)
		}

		baseName := filepath.Base(targetURL)
		if idx := strings.Index(baseName, "?"); idx != -1 {
			baseName = baseName[:idx]
		}

		destPath := filepath.Join(cg.StagingDir, baseName)
		out, err := os.Create(destPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create staging file: %v", err)
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to save file: %v", err)
		}

		status, err := cg.SandboxValidateSkillBook(destPath)
		if err != nil {
			_ = os.Remove(destPath)
			return nil, fmt.Errorf("sandbox validation failed for %s: %v", baseName, err)
		}

		log.Printf("📥 [WI-OS Package Manager] Successfully staged raw template: %s (Status: %s)", baseName, status)
		return []string{baseName}, nil
	}

	// GitHub Repository Archive conversion
	zipURL := targetURL
	if strings.Contains(targetURL, "github.com") && !strings.Contains(targetURL, ".zip") {
		targetURL = strings.TrimSuffix(targetURL, "/")
		parts := strings.Split(targetURL, "/")
		if len(parts) >= 5 {
			owner := parts[3]
			repo := parts[4]
			zipURL = fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/main.zip", owner, repo)
		}
	}

	resp, err := http.Get(zipURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch archive: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && strings.Contains(zipURL, "main.zip") {
		fallbackURL := strings.Replace(zipURL, "main.zip", "master.zip", 1)
		resp, err = http.Get(fallbackURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch fallback archive: %v", err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download repository archive (status %d)", resp.StatusCode)
	}

	zipBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read archive stream: %v", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to extract zip: %v", err)
	}

	var stagedFiles []string
	for _, file := range zipReader.File {
		if !file.FileInfo().IsDir() && filepath.Ext(file.Name) == ".md" {
			r, err := file.Open()
			if err != nil {
				continue
			}

			baseName := filepath.Base(file.Name)
			destPath := filepath.Join(cg.StagingDir, baseName)
			out, err := os.Create(destPath)
			if err != nil {
				r.Close()
				continue
			}

			_, err = io.Copy(out, r)
			out.Close()
			r.Close()

			_, err = cg.SandboxValidateSkillBook(destPath)
			if err != nil {
				_ = os.Remove(destPath)
				log.Printf("⚠️ [WI-OS Package Manager] Skipped malicious or invalid file inside archive: %s (%v)", baseName, err)
			} else {
				log.Printf("📥 [WI-OS Package Manager] Staged and verified template: %s", baseName)
				stagedFiles = append(stagedFiles, baseName)
			}
		}
	}

	if len(stagedFiles) == 0 {
		return nil, fmt.Errorf("no valid markdown skill books found in the repository")
	}

	return stagedFiles, nil
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

	_ = os.MkdirAll(filepath.Join("ai_skills", "unverified"), 0755)

	// Mock Diagnostic Test 1: Malicious file
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

	// Mock Diagnostic Test 2: Benign file
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

	err := cg.PromoteSkillBook("custom_market.md")
	if err == nil {
		fmt.Println("   Success: Skill Book verified and promoted to active catalog!")
		_ = os.Remove(filepath.Join("ai_skills", "custom_market.md"))
	} else {
		fmt.Printf("   Failed promotion: %v\n", err)
	}
	_ = os.Remove(bPath)

	// Test 3: Automated package downloader with a real URL
	fmt.Println("\nMock Test 3 (Package Installer) - Attempting direct raw download...")
	staged, err3 := cg.DownloadCommunitySkill("https://raw.githubusercontent.com/Ramcharan747/cursor-trajectory/main/README.md")
	if err3 != nil {
		fmt.Printf("   Note: Download failed or skipped (may be offline/unverified): %v\n", err3)
	} else {
		fmt.Printf("   Success: Downloaded and staged community files: %v\n", staged)
		for _, f := range staged {
			_ = os.Remove(filepath.Join(cg.StagingDir, f))
		}
	}
}
