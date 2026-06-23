{
  config,
  pkgs,
  lib,
  ...
}:
let
  eslint-wrapper = pkgs.writeShellApplication {
    name = "eslint-wrapper";
    runtimeInputs = [ pkgs.bun ];
    text = ''
      # Cap V8 heap at 1GB to keep the runner's OOM in check on
      # small-memory hosts (3–4GB). The default Node heap is ~1.7GB
      # and svelte-check / vite / paraglide will reliably OOM it.
      export NODE_OPTIONS="--max-old-space-size=1024"
      cd "${config.git.root}/frontend"
      bun install --frozen-lockfile
      bun run lint
    '';
  };
  svelte-check-wrapper = pkgs.writeShellApplication {
    name = "svelte-check-wrapper";
    runtimeInputs = [ pkgs.bun ];
    text = ''
      # Cap V8 heap at 1GB to keep the runner's OOM in check on
      # small-memory hosts (3–4GB). The default Node heap is ~1.7GB
      # and svelte-check / vite / paraglide will reliably OOM it.
      export NODE_OPTIONS="--max-old-space-size=1024"
      cd "${config.git.root}/frontend"
      bun install --frozen-lockfile
      bun run check
    '';
  };
  vitest-wrapper = pkgs.writeShellApplication {
    name = "vitest-wrapper";
    runtimeInputs = [ pkgs.bun ];
    text = ''
      # Cap V8 heap at 1GB to keep the runner's OOM in check on
      # small-memory hosts (3–4GB). The default Node heap is ~1.7GB
      # and svelte-check / vite / paraglide will reliably OOM it.
      export NODE_OPTIONS="--max-old-space-size=1024"
      cd "${config.git.root}/frontend"
      bun install --frozen-lockfile
      # Run tests only if test files exist, otherwise skip silently
      if find src -name "*.test.ts" -o -name "*.spec.ts" 2>/dev/null | grep -q .; then
        bun run test
      else
        echo "No test files found, skipping tests..."
        exit 0
      fi
    '';
  };
  frontend-build-wrapper = pkgs.writeShellApplication {
    name = "frontend-build-wrapper";
    runtimeInputs = [ pkgs.bun ];
    text = ''
      # Cap V8 heap at 1GB to keep the runner's OOM in check on
      # small-memory hosts (3–4GB). The default Node heap is ~1.7GB
      # and vite / paraglide will reliably OOM it.
      export NODE_OPTIONS="--max-old-space-size=1024"
      cd "${config.git.root}/frontend"
      bun install --frozen-lockfile
      bun run build
      mkdir -p "${config.git.root}/backend/cmd/openpost/public"
      touch "${config.git.root}/backend/cmd/openpost/public/.gitkeep"
    '';
  };
in
{
  # Bun language support
  languages.javascript = {
    enable = true;
    bun.enable = true;
  };

  # Scripts for frontend development
  scripts = {
    frontend-dev.exec = ''
      cd frontend && bun install && bun run dev
    '';

    frontend-build.exec = ''
      ${lib.getExe frontend-build-wrapper}
    '';

    frontend-test.exec = ''
      ${lib.getExe vitest-wrapper}
    '';

    frontend-check.exec = ''
      ${lib.getExe svelte-check-wrapper}
    '';

    frontend-lint.exec = ''
      ${lib.getExe eslint-wrapper}
    '';

    frontend-format.exec = ''
      cd frontend && bun run format
    '';
  };

  # Git hooks - all must pass to allow commits
  git-hooks.hooks = {
    # Lint check (prettier + eslint)
    eslint = {
      enable = true;
      entry = "${lib.getExe eslint-wrapper}";
      files = "\\.(js|ts|svelte)$";
      pass_filenames = false;
    };

    # Type check (svelte-check)
    svelte-check = {
      enable = true;
      entry = "${lib.getExe svelte-check-wrapper}";
      files = "\\.(ts|svelte)$";
      pass_filenames = false;
    };

    # Unit tests (vitest)
    vitest = {
      enable = true;
      entry = "${lib.getExe vitest-wrapper}";
      files = "^(frontend/(src|messages|static|assets)/|frontend/(package\\.json|bun\\.lock|vite\\.config\\.ts|svelte\\.config\\.js|vitest\\.config\\.[jt]s|tsconfig\\.json)|assets/|scripts/sync-assets\\.mjs)";
      pass_filenames = false;
    };

    # Production build catches Vite/Svelte compiler failures that
    # svelte-check and Vitest can miss.
    frontend-build = {
      enable = true;
      entry = "${lib.getExe frontend-build-wrapper}";
      files = "^(frontend/(src|messages|static|assets)/|frontend/(package\\.json|bun\\.lock|vite\\.config\\.ts|svelte\\.config\\.js)|assets/|scripts/sync-assets\\.mjs)";
      pass_filenames = false;
    };
  };
}
