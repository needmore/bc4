# This is a template for the Homebrew formula
# It should be placed in the homebrew-bc4 repository at Formula/bc4.rb

class Bc4 < Formula
  desc "A CLI tool for interacting with Basecamp 4"
  homepage "https://github.com/needmore/bc4"
  license "MIT"  # Update with actual license
  
  # Version will be updated by release process
  version "0.1.0"
  
  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/needmore/bc4/releases/download/v#{version}/bc4_#{version}_Darwin_arm64.tar.gz"
      sha256 "UPDATE_WITH_ACTUAL_SHA256"
    else
      url "https://github.com/needmore/bc4/releases/download/v#{version}/bc4_#{version}_Darwin_x86_64.tar.gz"
      sha256 "UPDATE_WITH_ACTUAL_SHA256"
    end
  end
  
  def install
    bin.install "bc4"
    
    # Install shell completions if they exist
    # bash_completion.install "completions/bc4.bash" if File.exist?("completions/bc4.bash")
    # zsh_completion.install "completions/_bc4" if File.exist?("completions/_bc4")
    # fish_completion.install "completions/bc4.fish" if File.exist?("completions/bc4.fish")
  end
  
  def post_install
    # Ensure config directory exists with correct permissions
    config_dir = "#{ENV["HOME"]}/.config/bc4"
    FileUtils.mkdir_p(config_dir) unless File.exist?(config_dir)
  end
  
  test do
    # Test version output
    assert_match "bc4 version", shell_output("#{bin}/bc4 --version")
    
    # Test help output
    assert_match "A CLI tool for interacting with Basecamp 4", shell_output("#{bin}/bc4 --help")
  end
end