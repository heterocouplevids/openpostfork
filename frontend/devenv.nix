{
  config,
  pkgs,
  lib,
  ...
}:
let
  eslint-wrapper = pkgs.writeShellApplication {
    name = "eslint-wrapper";
    runtimeInputs = [ pkgs.nodejs_22 pkgs.pnpm ];
    text = ''
      # Cap V8 heap at 1GB to keep the runner's OOM in check on
      # small-memory hosts (3–4GB). The default Node heap is ~1.7GB
      # and svelte-check / vite / paraglide will reliably OOM it.
      export NODE_OPTIONS="--max-old-space-size=1024"
      cd "${config.git.root}/frontend"
      pnpm install --frozen-lockfile
      pnpm --filter @openpost/web lint
    '';
  };
  svelte-check-wrapper = pkgs.writeShellApplication {
    name = "svelte-check-wrapper";
    runtimeInputs = [ pkgs.nodejs_22 pkgs.pnpm ];
    text = ''
      # Cap V8 heap at 1GB to keep the runner's OOM in check on
      # small-memory hosts (3–4GB). The default Node heap is ~1.7GB
      # and svelte-check / vite / paraglide will reliably OOM it.
      export NODE_OPTIONS="--max-old-space-size=1024"
      cd "${config.git.root}/frontend"
      pnpm install --frozen-lockfile
      pnpm --filter @openpost/web check
    '';
  };
  vitest-wrapper = pkgs.writeShellApplication {
    name = "vitest-wrapper";
    runtimeInputs = [ pkgs.nodejs_22 pkgs.pnpm ];
    text = ''
      # Cap V8 heap at 1GB to keep the runner's OOM in check on
      # small-memory hosts (3–4GB). The default Node heap is ~1.7GB
      # and svelte-check / vite / paraglide will reliably OOM it.
      export NODE_OPTIONS="--max-old-space-size=1024"
      cd "${config.git.root}/frontend"
      pnpm install --frozen-lockfile
      pnpm exec playwright install chromium
      # Run tests only if test files exist, otherwise skip silently
      if find src -name "*.test.ts" -o -name "*.spec.ts" 2>/dev/null | grep -q .; then
        pnpm --filter @openpost/web test
      else
        echo "No test files found, skipping tests..."
        exit 0
      fi
    '';
  };
  frontend-build-wrapper = pkgs.writeShellApplication {
    name = "frontend-build-wrapper";
    runtimeInputs = [ pkgs.nodejs_22 pkgs.pnpm ];
    text = ''
      # Cap V8 heap at 1GB to keep the runner's OOM in check on
      # small-memory hosts (3–4GB). The default Node heap is ~1.7GB
      # and vite / paraglide will reliably OOM it.
      export NODE_OPTIONS="--max-old-space-size=1024"
      cd "${config.git.root}/frontend"
      pnpm install --frozen-lockfile
      pnpm --filter @openpost/web build
      mkdir -p "${config.git.root}/backend/cmd/openpost/public"
      touch "${config.git.root}/backend/cmd/openpost/public/.gitkeep"
    '';
  };
in
{
  # JavaScript workspace support
  languages.javascript = {
    enable = true;
    pnpm.enable = true;
  };

  # Scripts for frontend development
  scripts = {
    frontend-dev.exec = ''
      pnpm install && pnpm --filter @openpost/web dev
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
      pnpm --filter @openpost/web format
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
      files = "^(frontend/(src|messages|static|assets)/|frontend/(package\\.json|vite\\.config\\.ts|svelte\\.config\\.js|vitest\\.config\\.[jt]s|tsconfig\\.json)|package\\.json|pnpm-lock\\.yaml|pnpm-workspace\\.yaml|turbo\\.json|assets/|scripts/sync-assets\\.mjs)";
      pass_filenames = false;
    };

    # Production build catches Vite/Svelte compiler failures that
    # svelte-check and Vitest can miss.
    frontend-build = {
      enable = true;
      entry = "${lib.getExe frontend-build-wrapper}";
      files = "^(frontend/(src|messages|static|assets)/|frontend/(package\\.json|vite\\.config\\.ts|svelte\\.config\\.js)|package\\.json|pnpm-lock\\.yaml|pnpm-workspace\\.yaml|turbo\\.json|assets/|scripts/sync-assets\\.mjs)";
      pass_filenames = false;
    };
  };
}
