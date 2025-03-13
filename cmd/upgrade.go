package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/lpsm-dev/gtoc/internal/logger"
	"github.com/spf13/cobra"
)

// GitHub API URLs
const (
	defaultGitHubAPIURL = "https://api.github.com/repos/lpsm-dev/gtoc/releases/latest"
	githubReleaseFormat = "https://github.com/lpsm-dev/gtoc/releases/download/%s/gtoc_%s_%s_%s"
)

// Release representa uma resposta da API do GitHub
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var (
	// Opções do comando upgrade
	forceUpgrade bool
	apiEndpoint  string
)

// upgradeCmd representa o comando upgrade
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade gtoc to the latest version",
	Long: `Upgrade checks GitHub for a newer version of gtoc and upgrades 
the current installation if a newer version is available.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Checking for updates", "current_version", Version)

		// Usar o endpoint padrão se não for especificado
		if apiEndpoint == "" {
			apiEndpoint = defaultGitHubAPIURL
		}

		// Buscar a última versão disponível
		release, err := getLatestRelease(apiEndpoint)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				logger.Warn("No releases found", "details", "You are using a development version or no releases are available yet")
				fmt.Println("No official releases found. You are using a development version or no releases are available yet.")
				return nil
			}
			logger.Error("Failed to check for updates", "error", err)
			return fmt.Errorf("failed to check for updates: %w", err)
		}

		// Mostrar informações da versão
		logger.Info("Latest version information", 
			"version", release.TagName, 
			"released_at", release.PublishedAt.Format("2006-01-02"))

		// Comparar versões
		if !shouldUpgrade(Version, release.TagName) && !forceUpgrade {
			fmt.Println("You already have the latest version installed!")
			return nil
		}

		// Prosseguir com a atualização
		fmt.Printf("Upgrading from version %s to %s\n", Version, release.TagName)
		
		// Determinar o arquivo correto a baixar
		downloadURL, err := getDownloadURL(release)
		if err != nil {
			logger.Error("Failed to determine download URL", "error", err)
			return err
		}

		logger.Debug("Downloading new version", "url", downloadURL)
		err = downloadAndInstall(downloadURL)
		if err != nil {
			logger.Error("Failed to download and install new version", "error", err)
			return err
		}

		fmt.Println("Successfully upgraded to version", release.TagName)
		return nil
	},
}

// getLatestRelease busca a informação da última release disponível no GitHub
func getLatestRelease(apiURL string) (*Release, error) {
	logger.Debug("Fetching latest release information", "url", apiURL)
	
	// Criar o HTTP client com timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Fazer a solicitação
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Adicionar User-Agent para evitar limitações de rate
	req.Header.Set("User-Agent", "gtoc-cli")

	// Executar a solicitação
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release information: %w", err)
	}
	defer resp.Body.Close()

	// Verificar status da resposta
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned non-OK status: %d", resp.StatusCode)
	}

	// Ler o corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Desserializar o JSON
	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release information: %w", err)
	}

	return &release, nil
}

// shouldUpgrade determina se deve atualizar com base nas versões
func shouldUpgrade(currentVersion, latestVersion string) bool {
	// Se a versão atual for "dev", sempre atualizar
	if currentVersion == "dev" {
		return true
	}

	// Remover 'v' inicial se existir
	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(latestVersion, "v")

	// Comparação simples de strings
	// Isso assume que as versões seguem o formato semântico e são comparáveis como strings
	return latest > current
}

// getDownloadURL determina a URL de download correta com base no sistema operacional
func getDownloadURL(release *Release) (string, error) {
	// Determinar a plataforma atual
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Mapear os valores do runtime para os nomes usados nos assets
	var os, arch string
	switch goos {
	case "darwin":
		os = "Darwin"
	case "linux":
		os = "Linux"
	case "windows":
		os = "Windows"
	default:
		return "", fmt.Errorf("unsupported operating system: %s", goos)
	}

	switch goarch {
	case "amd64":
		arch = "x86_64"
	case "386":
		arch = "i386"
	case "arm64":
		arch = "arm64"
	case "arm":
		arch = "arm"
	default:
		return "", fmt.Errorf("unsupported architecture: %s", goarch)
	}

	// Construir a URL de download
	tag := strings.TrimPrefix(release.TagName, "v")
	
	url := fmt.Sprintf(githubReleaseFormat, release.TagName, tag, os, arch)
	if goos == "windows" {
		url += ".exe"
	} else {
		url += ".tar.gz"
	}
	
	return url, nil
}

// downloadAndInstall baixa e instala a nova versão
func downloadAndInstall(url string) error {
	logger.Debug("Beginning download process", "url", url)

	// Criar o HTTP client com timeout
	client := &http.Client{
		Timeout: 5 * time.Minute, // Timeout mais longo para download de arquivos
	}

	// Fazer a solicitação
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	// Adicionar User-Agent
	req.Header.Set("User-Agent", "gtoc-cli")

	// Executar a solicitação
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// Verificar status da resposta
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download URL returned non-OK status: %d", resp.StatusCode)
	}

	// Determinar o caminho do executável atual
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine current executable path: %w", err)
	}
	logger.Debug("Current executable path", "path", execPath)

	// Para arquivos .tar.gz, precisaremos extrair após o download
	if strings.HasSuffix(url, ".tar.gz") {
		// Criar arquivo temporário para o download
		tempFile, err := os.CreateTemp("", "gtoc-*.tar.gz")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		tempPath := tempFile.Name()
		defer os.Remove(tempPath) // Limpar no final

		// Baixar para o arquivo temporário
		logger.Debug("Downloading to temporary file", "path", tempPath)
		_, err = io.Copy(tempFile, resp.Body)
		tempFile.Close()
		if err != nil {
			return fmt.Errorf("failed to write download to temp file: %w", err)
		}

		// Extrair e substituir o executável
		// Esta é uma implementação básica e pode precisar ser ajustada
		logger.Debug("Extracting archive and updating executable")
		
		// Criar diretório temporário para extração
		tempDir, err := os.MkdirTemp("", "gtoc-extract")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer os.RemoveAll(tempDir)
		
		// Extrair o arquivo
		extractCmd := fmt.Sprintf("tar -xzf %s -C %s", tempPath, tempDir)
		err = exec.Command("sh", "-c", extractCmd).Run()
		if err != nil {
			return fmt.Errorf("failed to extract archive: %w", err)
		}
		
		// Encontrar o binário extraído
		// Assumindo que o binário tem o nome 'gtoc'
		extractedBin := filepath.Join(tempDir, "gtoc")
		if _, err := os.Stat(extractedBin); os.IsNotExist(err) {
			// Tentar encontrar o binário em subdiretórios
			err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && (info.Name() == "gtoc" || info.Name() == "gtoc.exe") {
					extractedBin = path
					return filepath.SkipDir
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to locate extracted binary: %w", err)
			}
		}
		
		// Tornar o binário executável
		err = os.Chmod(extractedBin, 0755)
		if err != nil {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
		
		// Substituir o binário atual
		// Em sistemas Unix, não podemos substituir um binário em execução
		// Então criamos um script que substitui o binário após o término
		
		logger.Debug("Replacing current executable", "from", extractedBin, "to", execPath)
		err = os.Rename(extractedBin, execPath)
		if err != nil {
			// Se falhar devido ao binário estar em uso, tentar outra abordagem
			logger.Warn("Failed to replace executable directly", "error", err)
			logger.Info("Will attempt to replace on next run")
			
			// Criar um backup do binário atual
			backupPath := execPath + ".bak"
			err = os.Rename(execPath, backupPath)
			if err != nil {
				return fmt.Errorf("failed to backup current binary: %w", err)
			}
			
			// Copiar o novo binário para o local correto
			newFile, err := os.Create(execPath)
			if err != nil {
				// Restaurar o backup em caso de falha
				os.Rename(backupPath, execPath)
				return fmt.Errorf("failed to create new binary: %w", err)
			}
			defer newFile.Close()
			
			oldFile, err := os.Open(extractedBin)
			if err != nil {
				// Restaurar o backup em caso de falha
				os.Rename(backupPath, execPath)
				return fmt.Errorf("failed to open new binary: %w", err)
			}
			defer oldFile.Close()
			
			_, err = io.Copy(newFile, oldFile)
			if err != nil {
				// Restaurar o backup em caso de falha
				os.Rename(backupPath, execPath)
				return fmt.Errorf("failed to copy new binary: %w", err)
			}
			
			// Tornar o novo binário executável
			err = os.Chmod(execPath, 0755)
			if err != nil {
				// Restaurar o backup em caso de falha
				os.Rename(backupPath, execPath)
				return fmt.Errorf("failed to make new binary executable: %w", err)
			}
			
			// Remover o backup
			os.Remove(backupPath)
		}
		
	} else if strings.HasSuffix(url, ".exe") {
		// Para Windows, baixar diretamente sobre o binário atual
		// O Windows permite substituir arquivos em uso, mas requer
		// que os novos processos usem o novo arquivo
		
		// Criar uma cópia de backup
		backupPath := execPath + ".bak"
		err = os.Rename(execPath, backupPath)
		if err != nil {
			return fmt.Errorf("failed to backup current binary: %w", err)
		}
		
		// Baixar o novo binário
		newFile, err := os.Create(execPath)
		if err != nil {
			// Restaurar o backup
			os.Rename(backupPath, execPath)
			return fmt.Errorf("failed to create new binary: %w", err)
		}
		defer newFile.Close()
		
		_, err = io.Copy(newFile, resp.Body)
		if err != nil {
			// Restaurar o backup
			os.Rename(backupPath, execPath)
			return fmt.Errorf("failed to write new binary: %w", err)
		}
		
		// Remover o backup
		os.Remove(backupPath)
	}

	return nil
}

func init() {
	upgradeCmd.Flags().BoolVar(&forceUpgrade, "force", false, "Force upgrade even if the current version is the latest")
	upgradeCmd.Flags().StringVar(&apiEndpoint, "endpoint", "", "Specify a custom GitHub API endpoint")
	RootCmd.AddCommand(upgradeCmd)
} 