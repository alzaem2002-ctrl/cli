# البرومبت الشامل لإدارة المشاريع - Cursor Ultimate Prompt

## 🎯 نظرة عامة
دليل شامل لإدارة وتطوير المشاريع البرمجية باستخدام أفضل الممارسات والأدوات الحديثة.

---

## 📋 المراحل العشر التفصيلية

### المرحلة 1️⃣: إعداد البيئة الأولية (Initial Setup)

#### الأهداف:
- إنشاء بيئة عمل نظيفة ومعزولة
- تثبيت الأدوات الأساسية
- إعداد Git وإدارة النسخ

#### الأوامر:

```bash
# إنشاء مجلد المشروع
mkdir -p ~/projects/my-project
cd ~/projects/my-project

# إعداد Git
git init
git config user.name "Your Name"
git config user.email "your.email@example.com"

# إنشاء .gitignore
cat > .gitignore << 'EOF'
# Python
__pycache__/
*.py[cod]
*$py.class
*.so
.Python
env/
venv/
ENV/
*.egg-info/
dist/
build/

# IDEs
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Environment variables
.env
.env.local

# Logs
*.log
logs/

# Database
*.db
*.sqlite3

# Temporary files
tmp/
temp/
EOF

# إنشاء README.md
cat > README.md << 'EOF'
# Project Name

## Description
Brief description of your project

## Installation
```bash
pip install -r requirements.txt
```

## Usage
```bash
python main.py
```

## License
MIT License
EOF

# أول commit
git add .
git commit -m "Initial commit: project structure setup"
```

#### ملف Python: `setup_checker.py`

```python
#!/usr/bin/env python3
"""
Setup Checker - يتحقق من صحة إعداد البيئة
"""

import sys
import subprocess
import platform
from typing import Dict, List, Tuple

class SetupChecker:
    """فحص إعداد البيئة"""
    
    def __init__(self):
        self.results: List[Tuple[str, bool, str]] = []
    
    def check_python_version(self) -> bool:
        """التحقق من إصدار Python"""
        version = sys.version_info
        is_valid = version.major == 3 and version.minor >= 8
        
        self.results.append((
            "Python Version",
            is_valid,
            f"Python {version.major}.{version.minor}.{version.micro}"
        ))
        return is_valid
    
    def check_git(self) -> bool:
        """التحقق من تثبيت Git"""
        try:
            result = subprocess.run(
                ['git', '--version'],
                capture_output=True,
                text=True,
                timeout=5
            )
            is_valid = result.returncode == 0
            
            self.results.append((
                "Git Installation",
                is_valid,
                result.stdout.strip() if is_valid else "Not installed"
            ))
            return is_valid
        except Exception as e:
            self.results.append(("Git Installation", False, str(e)))
            return False
    
    def check_pip(self) -> bool:
        """التحقق من تثبيت pip"""
        try:
            result = subprocess.run(
                [sys.executable, '-m', 'pip', '--version'],
                capture_output=True,
                text=True,
                timeout=5
            )
            is_valid = result.returncode == 0
            
            self.results.append((
                "Pip Installation",
                is_valid,
                result.stdout.strip() if is_valid else "Not installed"
            ))
            return is_valid
        except Exception as e:
            self.results.append(("Pip Installation", False, str(e)))
            return False
    
    def check_virtual_env(self) -> bool:
        """التحقق من البيئة الافتراضية"""
        in_venv = sys.prefix != sys.base_prefix
        
        self.results.append((
            "Virtual Environment",
            in_venv,
            "Active" if in_venv else "Not active (recommended)"
        ))
        return True  # ليس إلزاميًا
    
    def print_results(self) -> None:
        """طباعة نتائج الفحص"""
        print("\n" + "="*60)
        print("Setup Verification Results".center(60))
        print("="*60 + "\n")
        
        for check_name, passed, details in self.results:
            status = "✅ PASS" if passed else "❌ FAIL"
            print(f"{status} | {check_name}")
            print(f"     Details: {details}\n")
        
        print("="*60)
        
        failed = [r for r in self.results if not r[1]]
        if failed:
            print(f"\n⚠️  {len(failed)} check(s) failed!")
            sys.exit(1)
        else:
            print("\n✅ All checks passed!")
            sys.exit(0)
    
    def run_all_checks(self) -> None:
        """تشغيل جميع الفحوصات"""
        print("🔍 Running setup verification checks...")
        
        self.check_python_version()
        self.check_git()
        self.check_pip()
        self.check_virtual_env()
        
        self.print_results()

if __name__ == "__main__":
    checker = SetupChecker()
    checker.run_all_checks()
```

---

### المرحلة 2️⃣: إنشاء البيئة الافتراضية (Virtual Environment)

#### الأهداف:
- عزل تبعيات المشروع
- تجنب تعارضات الحزم
- سهولة إعادة الإنتاج

#### الأوامر:

```bash
# إنشاء البيئة الافتراضية
python3 -m venv venv

# تفعيل البيئة (Linux/Mac)
source venv/bin/activate

# تفعيل البيئة (Windows)
# venv\Scripts\activate

# تحديث pip
pip install --upgrade pip setuptools wheel

# إنشاء ملف المتطلبات الأساسي
cat > requirements.txt << 'EOF'
# Core dependencies
requests>=2.31.0
python-dotenv>=1.0.0

# Development dependencies
pytest>=7.4.0
pytest-cov>=4.1.0
black>=23.0.0
flake8>=6.0.0
mypy>=1.5.0

# Utilities
colorama>=0.4.6
tqdm>=4.66.0
EOF

# تثبيت المتطلبات
pip install -r requirements.txt

# حفظ البيئة الحالية
pip freeze > requirements-lock.txt

# Commit التغييرات
git add requirements.txt requirements-lock.txt
git commit -m "feat: add project dependencies"
```

#### ملف Python: `venv_manager.py`

```python
#!/usr/bin/env python3
"""
Virtual Environment Manager - إدارة البيئات الافتراضية
"""

import os
import sys
import subprocess
import shutil
from pathlib import Path
from typing import Optional

class VenvManager:
    """مدير البيئات الافتراضية"""
    
    def __init__(self, venv_path: str = "venv"):
        self.venv_path = Path(venv_path)
        self.python_executable = sys.executable
    
    def create_venv(self, force: bool = False) -> bool:
        """إنشاء بيئة افتراضية جديدة"""
        try:
            if self.venv_path.exists():
                if not force:
                    print(f"❌ Virtual environment already exists at {self.venv_path}")
                    print("   Use --force to recreate")
                    return False
                
                print(f"🗑️  Removing existing venv at {self.venv_path}")
                shutil.rmtree(self.venv_path)
            
            print(f"🔨 Creating virtual environment at {self.venv_path}")
            subprocess.run(
                [self.python_executable, '-m', 'venv', str(self.venv_path)],
                check=True
            )
            
            print("✅ Virtual environment created successfully!")
            return True
            
        except subprocess.CalledProcessError as e:
            print(f"❌ Failed to create virtual environment: {e}")
            return False
        except Exception as e:
            print(f"❌ Unexpected error: {e}")
            return False
    
    def get_activation_command(self) -> str:
        """الحصول على أمر التفعيل المناسب للنظام"""
        if sys.platform == "win32":
            return str(self.venv_path / "Scripts" / "activate.bat")
        else:
            return f"source {self.venv_path}/bin/activate"
    
    def install_requirements(self, requirements_file: str = "requirements.txt") -> bool:
        """تثبيت المتطلبات من ملف"""
        try:
            if not Path(requirements_file).exists():
                print(f"❌ Requirements file not found: {requirements_file}")
                return False
            
            pip_executable = self._get_pip_executable()
            if not pip_executable:
                print("❌ Could not find pip executable in venv")
                return False
            
            print(f"📦 Installing packages from {requirements_file}")
            subprocess.run(
                [pip_executable, 'install', '-r', requirements_file],
                check=True
            )
            
            print("✅ Packages installed successfully!")
            return True
            
        except subprocess.CalledProcessError as e:
            print(f"❌ Failed to install packages: {e}")
            return False
        except Exception as e:
            print(f"❌ Unexpected error: {e}")
            return False
    
    def _get_pip_executable(self) -> Optional[Path]:
        """الحصول على مسار pip في البيئة الافتراضية"""
        if sys.platform == "win32":
            pip_path = self.venv_path / "Scripts" / "pip.exe"
        else:
            pip_path = self.venv_path / "bin" / "pip"
        
        return pip_path if pip_path.exists() else None
    
    def list_packages(self) -> bool:
        """عرض الحزم المثبتة"""
        try:
            pip_executable = self._get_pip_executable()
            if not pip_executable:
                print("❌ Could not find pip executable in venv")
                return False
            
            print("📦 Installed packages:\n")
            subprocess.run(
                [pip_executable, 'list'],
                check=True
            )
            return True
            
        except subprocess.CalledProcessError as e:
            print(f"❌ Failed to list packages: {e}")
            return False
    
    def freeze_requirements(self, output_file: str = "requirements-lock.txt") -> bool:
        """حفظ المتطلبات الحالية"""
        try:
            pip_executable = self._get_pip_executable()
            if not pip_executable:
                print("❌ Could not find pip executable in venv")
                return False
            
            print(f"💾 Freezing requirements to {output_file}")
            result = subprocess.run(
                [pip_executable, 'freeze'],
                capture_output=True,
                text=True,
                check=True
            )
            
            with open(output_file, 'w') as f:
                f.write(result.stdout)
            
            print(f"✅ Requirements saved to {output_file}")
            return True
            
        except subprocess.CalledProcessError as e:
            print(f"❌ Failed to freeze requirements: {e}")
            return False
        except Exception as e:
            print(f"❌ Unexpected error: {e}")
            return False

def main():
    """الدالة الرئيسية"""
    import argparse
    
    parser = argparse.ArgumentParser(description="Virtual Environment Manager")
    parser.add_argument('--create', action='store_true', help='Create virtual environment')
    parser.add_argument('--force', action='store_true', help='Force recreate if exists')
    parser.add_argument('--install', action='store_true', help='Install requirements')
    parser.add_argument('--list', action='store_true', help='List installed packages')
    parser.add_argument('--freeze', action='store_true', help='Freeze requirements')
    parser.add_argument('--venv-path', default='venv', help='Virtual environment path')
    
    args = parser.parse_args()
    
    manager = VenvManager(args.venv_path)
    
    if args.create:
        if manager.create_venv(force=args.force):
            print(f"\n💡 To activate the environment, run:")
            print(f"   {manager.get_activation_command()}")
    
    if args.install:
        manager.install_requirements()
    
    if args.list:
        manager.list_packages()
    
    if args.freeze:
        manager.freeze_requirements()

if __name__ == "__main__":
    main()
```

---

### المرحلة 3️⃣: هيكلة المشروع (Project Structure)

#### الأهداف:
- إنشاء هيكل منظم وقابل للتوسع
- فصل المسؤوليات
- سهولة الصيانة

#### الأوامر:

```bash
# إنشاء هيكل المجلدات
mkdir -p src/{core,utils,api,models,services}
mkdir -p tests/{unit,integration,e2e}
mkdir -p docs
mkdir -p config
mkdir -p scripts
mkdir -p data/{raw,processed}
mkdir -p logs

# إنشاء ملفات __init__.py
touch src/__init__.py
touch src/core/__init__.py
touch src/utils/__init__.py
touch src/api/__init__.py
touch src/models/__init__.py
touch src/services/__init__.py
touch tests/__init__.py
touch tests/unit/__init__.py
touch tests/integration/__init__.py

# Commit الهيكل
git add .
git commit -m "feat: create project structure"
```

#### ملف JSON: `project_structure.json`

```json
{
  "project_structure": {
    "version": "1.0.0",
    "description": "Standard project structure configuration",
    "directories": {
      "src": {
        "description": "Source code directory",
        "subdirectories": {
          "core": "Core business logic",
          "utils": "Utility functions and helpers",
          "api": "API endpoints and routes",
          "models": "Data models and schemas",
          "services": "Business services and external integrations"
        }
      },
      "tests": {
        "description": "Test directory",
        "subdirectories": {
          "unit": "Unit tests",
          "integration": "Integration tests",
          "e2e": "End-to-end tests"
        }
      },
      "docs": {
        "description": "Documentation files",
        "files": [
          "API.md",
          "ARCHITECTURE.md",
          "CONTRIBUTING.md"
        ]
      },
      "config": {
        "description": "Configuration files",
        "files": [
          "development.json",
          "production.json",
          "testing.json"
        ]
      },
      "scripts": {
        "description": "Utility scripts",
        "files": [
          "deploy.sh",
          "backup.sh",
          "migrate.py"
        ]
      },
      "data": {
        "description": "Data directory",
        "subdirectories": {
          "raw": "Raw data files",
          "processed": "Processed data files"
        }
      },
      "logs": {
        "description": "Log files directory"
      }
    },
    "root_files": [
      "README.md",
      "LICENSE",
      ".gitignore",
      "requirements.txt",
      "setup.py",
      ".env.example"
    ]
  }
}
```

#### ملف Python: `project_generator.py`

```python
#!/usr/bin/env python3
"""
Project Generator - مولد هيكل المشاريع التلقائي
"""

import os
import json
from pathlib import Path
from typing import Dict, List, Any

class ProjectGenerator:
    """مولد هيكل المشاريع"""
    
    def __init__(self, base_path: str = "."):
        self.base_path = Path(base_path)
        self.created_items: List[str] = []
        self.errors: List[str] = []
    
    def create_directory(self, path: Path, description: str = "") -> bool:
        """إنشاء مجلد"""
        try:
            path.mkdir(parents=True, exist_ok=True)
            self.created_items.append(f"📁 {path}")
            if description:
                print(f"✅ Created: {path} - {description}")
            return True
        except Exception as e:
            error_msg = f"Failed to create directory {path}: {e}"
            self.errors.append(error_msg)
            print(f"❌ {error_msg}")
            return False
    
    def create_file(self, path: Path, content: str = "", description: str = "") -> bool:
        """إنشاء ملف"""
        try:
            path.parent.mkdir(parents=True, exist_ok=True)
            
            if not path.exists():
                path.write_text(content)
                self.created_items.append(f"📄 {path}")
                if description:
                    print(f"✅ Created: {path} - {description}")
            else:
                print(f"⏭️  Skipped: {path} (already exists)")
            
            return True
        except Exception as e:
            error_msg = f"Failed to create file {path}: {e}"
            self.errors.append(error_msg)
            print(f"❌ {error_msg}")
            return False
    
    def generate_from_config(self, config_path: str) -> bool:
        """إنشاء هيكل من ملف JSON"""
        try:
            with open(config_path, 'r') as f:
                config = json.load(f)
            
            structure = config.get('project_structure', {})
            directories = structure.get('directories', {})
            
            print(f"\n🏗️  Generating project structure from {config_path}\n")
            
            # إنشاء المجلدات الرئيسية
            for dir_name, dir_config in directories.items():
                dir_path = self.base_path / dir_name
                description = dir_config.get('description', '')
                self.create_directory(dir_path, description)
                
                # إنشاء المجلدات الفرعية
                subdirs = dir_config.get('subdirectories', {})
                for subdir_name, subdir_desc in subdirs.items():
                    subdir_path = dir_path / subdir_name
                    self.create_directory(subdir_path, subdir_desc)
                    
                    # إنشاء __init__.py للمجلدات Python
                    if dir_name in ['src', 'tests'] or 'src' in str(dir_path):
                        init_file = subdir_path / '__init__.py'
                        self.create_file(init_file, '"""Package initialization"""')
                
                # إنشاء الملفات المحددة
                files = dir_config.get('files', [])
                for file_name in files:
                    file_path = dir_path / file_name
                    self.create_file(file_path, f"# {file_name}\n")
            
            # إنشاء الملفات الجذرية
            root_files = structure.get('root_files', [])
            for file_name in root_files:
                file_path = self.base_path / file_name
                if not file_path.exists():
                    content = self._get_default_content(file_name)
                    self.create_file(file_path, content)
            
            return True
            
        except FileNotFoundError:
            print(f"❌ Config file not found: {config_path}")
            return False
        except json.JSONDecodeError as e:
            print(f"❌ Invalid JSON in config file: {e}")
            return False
        except Exception as e:
            print(f"❌ Unexpected error: {e}")
            return False
    
    def _get_default_content(self, filename: str) -> str:
        """الحصول على المحتوى الافتراضي للملفات"""
        templates = {
            'README.md': '''# Project Name

## Description
Add your project description here

## Installation
```bash
pip install -r requirements.txt
```

## Usage
```bash
python main.py
```

## Testing
```bash
pytest tests/
```

## License
MIT
''',
            'LICENSE': '''MIT License

Copyright (c) 2024

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction...
''',
            '.env.example': '''# Environment Variables Example
# Copy this file to .env and fill in your values

# Database
DATABASE_URL=postgresql://user:password@localhost/dbname

# API Keys
API_KEY=your_api_key_here
SECRET_KEY=your_secret_key_here

# Environment
ENV=development
DEBUG=True
''',
            'setup.py': '''from setuptools import setup, find_packages

setup(
    name="project-name",
    version="0.1.0",
    packages=find_packages(),
    install_requires=[
        # Add your dependencies here
    ],
)
'''
        }
        
        return templates.get(filename, f"# {filename}\n")
    
    def print_summary(self) -> None:
        """طباعة ملخص العملية"""
        print("\n" + "="*60)
        print("Project Generation Summary".center(60))
        print("="*60 + "\n")
        
        print(f"✅ Created {len(self.created_items)} items")
        
        if self.errors:
            print(f"\n❌ Encountered {len(self.errors)} errors:")
            for error in self.errors:
                print(f"   • {error}")
        else:
            print("\n🎉 All items created successfully!")
        
        print("\n" + "="*60)

def main():
    """الدالة الرئيسية"""
    import argparse
    
    parser = argparse.ArgumentParser(description="Project Structure Generator")
    parser.add_argument('--config', default='project_structure.json',
                       help='Path to configuration JSON file')
    parser.add_argument('--base-path', default='.',
                       help='Base path for project generation')
    
    args = parser.parse_args()
    
    generator = ProjectGenerator(args.base_path)
    
    if generator.generate_from_config(args.config):
        generator.print_summary()
        return 0
    else:
        print("\n❌ Project generation failed")
        return 1

if __name__ == "__main__":
    exit(main())
```

---

### المرحلة 4️⃣: إعداد الإعدادات والمتغيرات البيئية (Configuration)

#### الأهداف:
- إدارة الإعدادات بشكل مركزي
- فصل الإعدادات حسب البيئة
- حماية البيانات الحساسة

#### الأوامر:

```bash
# إنشاء ملف .env.example
cat > .env.example << 'EOF'
# Application Settings
APP_NAME=MyApplication
APP_VERSION=1.0.0
ENV=development
DEBUG=True

# Database Configuration
DATABASE_URL=postgresql://user:password@localhost:5432/dbname
DATABASE_POOL_SIZE=10
DATABASE_TIMEOUT=30

# API Configuration
API_HOST=0.0.0.0
API_PORT=8000
API_PREFIX=/api/v1

# Security
SECRET_KEY=your-secret-key-here
JWT_SECRET=your-jwt-secret-here
JWT_EXPIRATION=3600

# External Services
REDIS_URL=redis://localhost:6379/0
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-password

# Logging
LOG_LEVEL=INFO
LOG_FILE=logs/app.log

# Feature Flags
FEATURE_NEW_UI=false
FEATURE_BETA_API=false
EOF

# نسخ الملف للاستخدام المحلي
cp .env.example .env

# إنشاء ملفات الإعدادات
cat > config/development.json << 'EOF'
{
  "environment": "development",
  "debug": true,
  "database": {
    "host": "localhost",
    "port": 5432,
    "name": "dev_db"
  },
  "cache": {
    "enabled": false
  },
  "logging": {
    "level": "DEBUG",
    "format": "detailed"
  }
}
EOF

cat > config/production.json << 'EOF'
{
  "environment": "production",
  "debug": false,
  "database": {
    "host": "prod-db.example.com",
    "port": 5432,
    "name": "prod_db",
    "ssl": true
  },
  "cache": {
    "enabled": true,
    "ttl": 3600
  },
  "logging": {
    "level": "WARNING",
    "format": "json"
  }
}
EOF

# Commit
git add .env.example config/
git commit -m "feat: add configuration files"
```

#### ملف Python: `config_manager.py`

```python
#!/usr/bin/env python3
"""
Configuration Manager - مدير الإعدادات المركزي
"""

import os
import json
from pathlib import Path
from typing import Any, Dict, Optional
from dotenv import load_dotenv

class ConfigError(Exception):
    """استثناء خاص بأخطاء الإعدادات"""
    pass

class Config:
    """فئة الإعدادات"""
    
    def __init__(self, env: str = None):
        self.env = env or os.getenv('ENV', 'development')
        self._config: Dict[str, Any] = {}
        self._loaded = False
        
        # تحميل المتغيرات البيئية
        load_dotenv()
        
        # تحميل الإعدادات
        self._load_config()
    
    def _load_config(self) -> None:
        """تحميل ملف الإعدادات المناسب"""
        try:
            config_file = Path('config') / f'{self.env}.json'
            
            if not config_file.exists():
                raise ConfigError(f"Configuration file not found: {config_file}")
            
            with open(config_file, 'r') as f:
                self._config = json.load(f)
            
            self._loaded = True
            print(f"✅ Configuration loaded for environment: {self.env}")
            
        except json.JSONDecodeError as e:
            raise ConfigError(f"Invalid JSON in configuration file: {e}")
        except Exception as e:
            raise ConfigError(f"Failed to load configuration: {e}")
    
    def get(self, key: str, default: Any = None) -> Any:
        """الحصول على قيمة إعداد"""
        if not self._loaded:
            raise ConfigError("Configuration not loaded")
        
        # البحث في المتغيرات البيئية أولاً
        env_value = os.getenv(key.upper())
        if env_value is not None:
            return self._cast_value(env_value)
        
        # البحث في ملف الإعدادات
        keys = key.split('.')
        value = self._config
        
        for k in keys:
            if isinstance(value, dict):
                value = value.get(k)
            else:
                return default
        
        return value if value is not None else default
    
    def _cast_value(self, value: str) -> Any:
        """تحويل القيم النصية إلى الأنواع المناسبة"""
        # Boolean
        if value.lower() in ('true', 'yes', '1'):
            return True
        if value.lower() in ('false', 'no', '0'):
            return False
        
        # Integer
        try:
            return int(value)
        except ValueError:
            pass
        
        # Float
        try:
            return float(value)
        except ValueError:
            pass
        
        # String
        return value
    
    def get_required(self, key: str) -> Any:
        """الحصول على قيمة إجبارية"""
        value = self.get(key)
        if value is None:
            raise ConfigError(f"Required configuration key not found: {key}")
        return value
    
    def set(self, key: str, value: Any) -> None:
        """تعيين قيمة إعداد (للاختبار فقط)"""
        keys = key.split('.')
        config = self._config
        
        for k in keys[:-1]:
            if k not in config:
                config[k] = {}
            config = config[k]
        
        config[keys[-1]] = value
    
    def to_dict(self) -> Dict[str, Any]:
        """تحويل الإعدادات إلى قاموس"""
        return self._config.copy()
    
    def validate(self, required_keys: list) -> bool:
        """التحقق من وجود المفاتيح المطلوبة"""
        missing_keys = []
        
        for key in required_keys:
            try:
                self.get_required(key)
            except ConfigError:
                missing_keys.append(key)
        
        if missing_keys:
            raise ConfigError(f"Missing required configuration keys: {missing_keys}")
        
        return True

class DatabaseConfig:
    """إعدادات قاعدة البيانات"""
    
    def __init__(self, config: Config):
        self.config = config
    
    @property
    def url(self) -> str:
        """رابط قاعدة البيانات"""
        url = self.config.get('DATABASE_URL')
        if url:
            return url
        
        # بناء الرابط من المكونات
        host = self.config.get('database.host', 'localhost')
        port = self.config.get('database.port', 5432)
        name = self.config.get('database.name', 'mydb')
        user = self.config.get('database.user', 'user')
        password = self.config.get('database.password', 'password')
        
        return f"postgresql://{user}:{password}@{host}:{port}/{name}"
    
    @property
    def pool_size(self) -> int:
        """حجم Pool الاتصالات"""
        return self.config.get('DATABASE_POOL_SIZE', 10)
    
    @property
    def timeout(self) -> int:
        """مهلة الاتصال"""
        return self.config.get('DATABASE_TIMEOUT', 30)

class APIConfig:
    """إعدادات API"""
    
    def __init__(self, config: Config):
        self.config = config
    
    @property
    def host(self) -> str:
        return self.config.get('API_HOST', '0.0.0.0')
    
    @property
    def port(self) -> int:
        return self.config.get('API_PORT', 8000)
    
    @property
    def prefix(self) -> str:
        return self.config.get('API_PREFIX', '/api/v1')
    
    @property
    def url(self) -> str:
        return f"http://{self.host}:{self.port}{self.prefix}"

class AppConfig:
    """تجميع جميع الإعدادات"""
    
    def __init__(self, env: str = None):
        self._config = Config(env)
        self.database = DatabaseConfig(self._config)
        self.api = APIConfig(self._config)
    
    @property
    def name(self) -> str:
        return self._config.get('APP_NAME', 'MyApp')
    
    @property
    def version(self) -> str:
        return self._config.get('APP_VERSION', '1.0.0')
    
    @property
    def debug(self) -> bool:
        return self._config.get('DEBUG', False)
    
    @property
    def environment(self) -> str:
        return self._config.env
    
    def get(self, key: str, default: Any = None) -> Any:
        return self._config.get(key, default)

# مثال للاستخدام
if __name__ == "__main__":
    try:
        # تحميل الإعدادات
        config = AppConfig()
        
        print(f"🚀 Application: {config.name} v{config.version}")
        print(f"🌍 Environment: {config.environment}")
        print(f"🐛 Debug Mode: {config.debug}")
        print(f"🗄️  Database URL: {config.database.url}")
        print(f"🌐 API URL: {config.api.url}")
        
    except ConfigError as e:
        print(f"❌ Configuration Error: {e}")
        exit(1)
```

---

### المرحلة 5️⃣: معالجة الأخطاء والسجلات (Error Handling & Logging)

#### الأهداف:
- معالجة شاملة للأخطاء
- تسجيل مفصل للأحداث
- تتبع الأخطاء وحلها

#### الأوامر:

```bash
# إنشاء مجلد السجلات
mkdir -p logs

# إضافة إلى .gitignore
echo "logs/*.log" >> .gitignore

# Commit
git add .gitignore
git commit -m "feat: add logging configuration"
```

#### ملف Python: `logger.py`

```python
#!/usr/bin/env python3
"""
Advanced Logging System - نظام تسجيل متقدم
"""

import logging
import sys
from pathlib import Path
from typing import Optional
from datetime import datetime
import traceback
import json

class ColoredFormatter(logging.Formatter):
    """Formatter ملون للطباعة في Console"""
    
    COLORS = {
        'DEBUG': '\033[36m',      # Cyan
        'INFO': '\033[32m',       # Green
        'WARNING': '\033[33m',    # Yellow
        'ERROR': '\033[31m',      # Red
        'CRITICAL': '\033[35m',   # Magenta
        'RESET': '\033[0m'        # Reset
    }
    
    def format(self, record: logging.LogRecord) -> str:
        """تنسيق الرسالة مع الألوان"""
        color = self.COLORS.get(record.levelname, self.COLORS['RESET'])
        reset = self.COLORS['RESET']
        
        # تلوين اسم المستوى فقط
        record.levelname = f"{color}{record.levelname}{reset}"
        
        return super().format(record)

class JSONFormatter(logging.Formatter):
    """Formatter لإخراج JSON"""
    
    def format(self, record: logging.LogRecord) -> str:
        """تنسيق الرسالة كـ JSON"""
        log_data = {
            'timestamp': datetime.fromtimestamp(record.created).isoformat(),
            'level': record.levelname,
            'logger': record.name,
            'message': record.getMessage(),
            'module': record.module,
            'function': record.funcName,
            'line': record.lineno
        }
        
        if record.exc_info:
            log_data['exception'] = self.formatException(record.exc_info)
        
        return json.dumps(log_data, ensure_ascii=False)

class Logger:
    """مدير السجلات المتقدم"""
    
    def __init__(
        self,
        name: str = 'app',
        level: str = 'INFO',
        log_file: Optional[str] = None,
        json_format: bool = False,
        console: bool = True
    ):
        self.logger = logging.getLogger(name)
        self.logger.setLevel(getattr(logging, level.upper()))
        
        # تجنب تكرار Handlers
        if self.logger.handlers:
            return
        
        # Console Handler
        if console:
            console_handler = logging.StreamHandler(sys.stdout)
            console_handler.setLevel(logging.DEBUG)
            
            if json_format:
                console_formatter = JSONFormatter()
            else:
                console_formatter = ColoredFormatter(
                    '%(asctime)s | %(levelname)s | %(name)s | %(message)s',
                    datefmt='%Y-%m-%d %H:%M:%S'
                )
            
            console_handler.setFormatter(console_formatter)
            self.logger.addHandler(console_handler)
        
        # File Handler
        if log_file:
            log_path = Path(log_file)
            log_path.parent.mkdir(parents=True, exist_ok=True)
            
            file_handler = logging.FileHandler(log_file, encoding='utf-8')
            file_handler.setLevel(logging.DEBUG)
            
            if json_format:
                file_formatter = JSONFormatter()
            else:
                file_formatter = logging.Formatter(
                    '%(asctime)s | %(levelname)-8s | %(name)s | %(funcName)s:%(lineno)d | %(message)s',
                    datefmt='%Y-%m-%d %H:%M:%S'
                )
            
            file_handler.setFormatter(file_formatter)
            self.logger.addHandler(file_handler)
    
    def debug(self, message: str, **kwargs) -> None:
        """رسالة تصحيح"""
        self.logger.debug(message, extra=kwargs)
    
    def info(self, message: str, **kwargs) -> None:
        """رسالة معلومات"""
        self.logger.info(message, extra=kwargs)
    
    def warning(self, message: str, **kwargs) -> None:
        """رسالة تحذير"""
        self.logger.warning(message, extra=kwargs)
    
    def error(self, message: str, exc_info: bool = False, **kwargs) -> None:
        """رسالة خطأ"""
        self.logger.error(message, exc_info=exc_info, extra=kwargs)
    
    def critical(self, message: str, exc_info: bool = True, **kwargs) -> None:
        """رسالة حرجة"""
        self.logger.critical(message, exc_info=exc_info, extra=kwargs)
    
    def exception(self, message: str, **kwargs) -> None:
        """تسجيل استثناء"""
        self.logger.exception(message, extra=kwargs)

# مثال للاستخدام
if __name__ == "__main__":
    # إنشاء logger
    logger = Logger(
        name='demo',
        level='DEBUG',
        log_file='logs/app.log',
        json_format=False
    )
    
    logger.info("🚀 Application started")
    logger.debug("Debug information")
    logger.warning("⚠️ This is a warning")
    
    try:
        result = 10 / 0
    except Exception as e:
        logger.exception("❌ An error occurred")
    
    logger.info("✅ Application finished")
```

#### ملف Python: `error_handler.py`

```python
#!/usr/bin/env python3
"""
Error Handler - معالج أخطاء شامل
"""

import sys
import traceback
from typing import Type, Callable, Optional, Any
from functools import wraps
from dataclasses import dataclass
from datetime import datetime

@dataclass
class ErrorContext:
    """سياق الخطأ"""
    timestamp: datetime
    error_type: str
    error_message: str
    traceback: str
    function_name: str
    file_name: str
    line_number: int
    
    def to_dict(self) -> dict:
        """تحويل إلى قاموس"""
        return {
            'timestamp': self.timestamp.isoformat(),
            'error_type': self.error_type,
            'error_message': self.error_message,
            'traceback': self.traceback,
            'location': {
                'function': self.function_name,
                'file': self.file_name,
                'line': self.line_number
            }
        }

class ErrorHandler:
    """معالج الأخطاء الرئيسي"""
    
    def __init__(self, logger=None):
        self.logger = logger
        self.error_callbacks = []
    
    def register_callback(self, callback: Callable) -> None:
        """تسجيل callback للأخطاء"""
        self.error_callbacks.append(callback)
    
    def handle_error(
        self,
        exception: Exception,
        context: Optional[ErrorContext] = None,
        reraise: bool = False
    ) -> None:
        """معالجة الخطأ"""
        if context is None:
            context = self._create_context(exception)
        
        # تسجيل الخطأ
        if self.logger:
            self.logger.error(
                f"Error occurred: {context.error_message}",
                exc_info=True
            )
        else:
            print(f"❌ Error: {context.error_message}", file=sys.stderr)
            print(f"   Type: {context.error_type}", file=sys.stderr)
            print(f"   Location: {context.file_name}:{context.line_number}", file=sys.stderr)
        
        # استدعاء Callbacks
        for callback in self.error_callbacks:
            try:
                callback(context)
            except Exception as e:
                print(f"Error in callback: {e}", file=sys.stderr)
        
        # إعادة رفع الخطأ إذا طلب
        if reraise:
            raise exception
    
    def _create_context(self, exception: Exception) -> ErrorContext:
        """إنشاء سياق الخطأ"""
        tb = sys.exc_info()[2]
        tb_info = traceback.extract_tb(tb)[-1] if tb else None
        
        return ErrorContext(
            timestamp=datetime.now(),
            error_type=type(exception).__name__,
            error_message=str(exception),
            traceback=traceback.format_exc(),
            function_name=tb_info.name if tb_info else 'unknown',
            file_name=tb_info.filename if tb_info else 'unknown',
            line_number=tb_info.lineno if tb_info else 0
        )
    
    def decorator(
        self,
        reraise: bool = False,
        default_return: Any = None
    ) -> Callable:
        """Decorator لمعالجة الأخطاء"""
        def wrapper(func: Callable) -> Callable:
            @wraps(func)
            def inner(*args, **kwargs):
                try:
                    return func(*args, **kwargs)
                except Exception as e:
                    self.handle_error(e, reraise=reraise)
                    return default_return
            return inner
        return wrapper

class RetryHandler:
    """معالج إعادة المحاولة"""
    
    @staticmethod
    def retry(
        max_attempts: int = 3,
        delay: float = 1.0,
        backoff: float = 2.0,
        exceptions: tuple = (Exception,)
    ) -> Callable:
        """Decorator لإعادة المحاولة عند الفشل"""
        def decorator(func: Callable) -> Callable:
            @wraps(func)
            def wrapper(*args, **kwargs):
                import time
                
                attempt = 1
                current_delay = delay
                
                while attempt <= max_attempts:
                    try:
                        return func(*args, **kwargs)
                    except exceptions as e:
                        if attempt == max_attempts:
                            raise
                        
                        print(f"⚠️ Attempt {attempt}/{max_attempts} failed: {e}")
                        print(f"   Retrying in {current_delay:.1f} seconds...")
                        
                        time.sleep(current_delay)
                        current_delay *= backoff
                        attempt += 1
                
                raise RuntimeError("Max retry attempts reached")
            
            return wrapper
        return decorator

# أمثلة للاستخدام
if __name__ == "__main__":
    from logger import Logger
    
    # إعداد Logger
    logger = Logger(name='error_demo', level='DEBUG')
    
    # إعداد Error Handler
    error_handler = ErrorHandler(logger)
    
    # مثال 1: استخدام decorator
    @error_handler.decorator(reraise=False, default_return=None)
    def risky_function():
        """دالة قد تفشل"""
        print("Executing risky function...")
        result = 10 / 0  # سيسبب خطأ
        return result
    
    print("\n=== Test 1: Error Handler Decorator ===")
    result = risky_function()
    print(f"Result: {result}")
    
    # مثال 2: إعادة المحاولة
    @RetryHandler.retry(max_attempts=3, delay=0.5, exceptions=(ValueError,))
    def unstable_function(succeed_on_attempt: int = 3):
        """دالة غير مستقرة"""
        import random
        attempt = getattr(unstable_function, '_attempt', 0) + 1
        unstable_function._attempt = attempt
        
        print(f"  Attempt {attempt}")
        
        if attempt < succeed_on_attempt:
            raise ValueError(f"Failed on attempt {attempt}")
        
        return "Success!"
    
    print("\n=== Test 2: Retry Handler ===")
    try:
        result = unstable_function(succeed_on_attempt=2)
        print(f"✅ {result}")
    except Exception as e:
        print(f"❌ Final failure: {e}")
```

---

### المرحلة 6️⃣: إدارة قاعدة البيانات (Database Management)

#### الأهداف:
- إدارة اتصالات قاعدة البيانات
- معالجة العمليات بشكل آمن
- إدارة Migrations

#### الأوامر:

```bash
# تثبيت المكتبات المطلوبة
pip install sqlalchemy psycopg2-binary alembic

# تحديث requirements.txt
pip freeze | grep -E "(SQLAlchemy|psycopg2|alembic)" >> requirements.txt

# إنشاء مجلد migrations
mkdir -p migrations

# Commit
git add requirements.txt
git commit -m "feat: add database dependencies"
```

#### ملف Python: `database.py`

```python
#!/usr/bin/env python3
"""
Database Manager - مدير قاعدة البيانات
"""

from typing import Optional, Any, Dict, List
from contextlib import contextmanager
import time

try:
    from sqlalchemy import create_engine, event, pool
    from sqlalchemy.ext.declarative import declarative_base
    from sqlalchemy.orm import sessionmaker, Session
    from sqlalchemy.exc import SQLAlchemyError, OperationalError
    SQLALCHEMY_AVAILABLE = True
except ImportError:
    SQLALCHEMY_AVAILABLE = False
    print("⚠️ SQLAlchemy not installed. Install with: pip install sqlalchemy")

# Base للموديلات
Base = declarative_base() if SQLALCHEMY_AVAILABLE else None

class DatabaseManager:
    """مدير قاعدة البيانات الشامل"""
    
    def __init__(
        self,
        database_url: str,
        echo: bool = False,
        pool_size: int = 10,
        max_overflow: int = 20,
        pool_timeout: int = 30,
        pool_recycle: int = 3600
    ):
        if not SQLALCHEMY_AVAILABLE:
            raise ImportError("SQLAlchemy is required for DatabaseManager")
        
        self.database_url = database_url
        self.engine = None
        self.SessionLocal = None
        self._connected = False
        
        # إنشاء Engine
        self.engine = create_engine(
            database_url,
            echo=echo,
            poolclass=pool.QueuePool,
            pool_size=pool_size,
            max_overflow=max_overflow,
            pool_timeout=pool_timeout,
            pool_recycle=pool_recycle,
            pool_pre_ping=True  # للتحقق من صحة الاتصالات
        )
        
        # إنشاء Session Factory
        self.SessionLocal = sessionmaker(
            autocommit=False,
            autoflush=False,
            bind=self.engine
        )
        
        # إضافة Event Listeners
        self._setup_event_listeners()
    
    def _setup_event_listeners(self) -> None:
        """إعداد Event Listeners لمراقبة الأداء"""
        @event.listens_for(self.engine, "before_cursor_execute")
        def before_cursor_execute(conn, cursor, statement, parameters, context, executemany):
            context._query_start_time = time.time()
        
        @event.listens_for(self.engine, "after_cursor_execute")
        def after_cursor_execute(conn, cursor, statement, parameters, context, executemany):
            total_time = time.time() - context._query_start_time
            if total_time > 1.0:  # تحذير للاستعلامات البطيئة
                print(f"⚠️ Slow query detected ({total_time:.2f}s): {statement[:100]}")
    
    def connect(self) -> bool:
        """الاتصال بقاعدة البيانات"""
        try:
            # اختبار الاتصال
            with self.engine.connect() as connection:
                connection.execute("SELECT 1")
            
            self._connected = True
            print("✅ Database connected successfully")
            return True
            
        except OperationalError as e:
            print(f"❌ Failed to connect to database: {e}")
            self._connected = False
            return False
        except Exception as e:
            print(f"❌ Unexpected error during connection: {e}")
            self._connected = False
            return False
    
    def disconnect(self) -> None:
        """قطع الاتصال"""
        if self.engine:
            self.engine.dispose()
            self._connected = False
            print("✅ Database disconnected")
    
    def create_tables(self) -> bool:
        """إنشاء جميع الجداول"""
        try:
            Base.metadata.create_all(bind=self.engine)
            print("✅ Tables created successfully")
            return True
        except Exception as e:
            print(f"❌ Failed to create tables: {e}")
            return False
    
    def drop_tables(self) -> bool:
        """حذف جميع الجداول (خطر!)"""
        try:
            Base.metadata.drop_all(bind=self.engine)
            print("✅ Tables dropped successfully")
            return True
        except Exception as e:
            print(f"❌ Failed to drop tables: {e}")
            return False
    
    @contextmanager
    def get_session(self):
        """الحصول على Session مع Context Manager"""
        session = self.SessionLocal()
        try:
            yield session
            session.commit()
        except Exception as e:
            session.rollback()
            print(f"❌ Session error: {e}")
            raise
        finally:
            session.close()
    
    def execute_query(self, query: str, params: Dict = None) -> List[Any]:
        """تنفيذ استعلام مباشر"""
        try:
            with self.engine.connect() as connection:
                result = connection.execute(query, params or {})
                return result.fetchall()
        except Exception as e:
            print(f"❌ Query execution failed: {e}")
            raise
    
    def health_check(self) -> Dict[str, Any]:
        """فحص صحة قاعدة البيانات"""
        health_status = {
            'connected': False,
            'pool_size': 0,
            'pool_checked_in': 0,
            'pool_checked_out': 0,
            'pool_overflow': 0,
            'response_time': None
        }
        
        try:
            start_time = time.time()
            
            with self.engine.connect() as connection:
                connection.execute("SELECT 1")
            
            health_status['response_time'] = time.time() - start_time
            health_status['connected'] = True
            
            # معلومات Pool
            pool_status = self.engine.pool.status()
            health_status['pool_size'] = self.engine.pool.size()
            health_status['pool_checked_in'] = self.engine.pool.checkedin()
            health_status['pool_checked_out'] = self.engine.pool.checkedout()
            health_status['pool_overflow'] = self.engine.pool.overflow()
            
        except Exception as e:
            health_status['error'] = str(e)
        
        return health_status

# مثال للاستخدام
if __name__ == "__main__":
    # إنشاء مدير قاعدة البيانات
    db_url = "sqlite:///./test.db"  # أو استخدم PostgreSQL URL
    db = DatabaseManager(database_url=db_url, echo=True)
    
    # الاتصال
    if db.connect():
        # فحص الصحة
        health = db.health_check()
        print(f"\n📊 Database Health:")
        for key, value in health.items():
            print(f"   {key}: {value}")
        
        # استخدام Session
        with db.get_session() as session:
            # أداء عمليات قاعدة البيانات هنا
            pass
        
        # قطع الاتصال
        db.disconnect()
```

---

### المرحلة 7️⃣: بناء API (API Development)

#### الأهداف:
- إنشاء RESTful API
- معالجة الطلبات والاستجابات
- توثيق API

#### الأوامر:

```bash
# تثبيت FastAPI وأدواته
pip install fastapi uvicorn[standard] pydantic[email]

# تحديث requirements.txt
pip freeze | grep -E "(fastapi|uvicorn|pydantic)" >> requirements.txt

# Commit
git add requirements.txt
git commit -m "feat: add API dependencies"
```

#### ملف Python: `api_server.py`

```python
#!/usr/bin/env python3
"""
FastAPI Server - خادم API متقدم
"""

from typing import Optional, List, Dict, Any
from datetime import datetime
from enum import Enum

try:
    from fastapi import FastAPI, HTTPException, Depends, status, Query
    from fastapi.middleware.cors import CORSMiddleware
    from fastapi.responses import JSONResponse
    from pydantic import BaseModel, Field, validator
    import uvicorn
    FASTAPI_AVAILABLE = True
except ImportError:
    FASTAPI_AVAILABLE = False
    print("⚠️ FastAPI not installed. Install with: pip install fastapi uvicorn")

if FASTAPI_AVAILABLE:
    # إنشاء التطبيق
    app = FastAPI(
        title="My API",
        description="API Documentation",
        version="1.0.0",
        docs_url="/docs",
        redoc_url="/redoc"
    )
    
    # إعداد CORS
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],  # في الإنتاج، حدد النطاقات المسموح بها
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    # Models
    class StatusEnum(str, Enum):
        """حالات العنصر"""
        ACTIVE = "active"
        INACTIVE = "inactive"
        PENDING = "pending"
    
    class ItemBase(BaseModel):
        """نموذج العنصر الأساسي"""
        name: str = Field(..., min_length=1, max_length=100)
        description: Optional[str] = Field(None, max_length=500)
        price: float = Field(..., gt=0)
        status: StatusEnum = StatusEnum.ACTIVE
        tags: List[str] = Field(default_factory=list)
        
        @validator('price')
        def validate_price(cls, v):
            if v < 0:
                raise ValueError('Price must be positive')
            return round(v, 2)
        
        class Config:
            schema_extra = {
                "example": {
                    "name": "Sample Item",
                    "description": "A sample item",
                    "price": 29.99,
                    "status": "active",
                    "tags": ["new", "featured"]
                }
            }
    
    class ItemCreate(ItemBase):
        """نموذج إنشاء عنصر"""
        pass
    
    class ItemUpdate(BaseModel):
        """نموذج تحديث عنصر"""
        name: Optional[str] = Field(None, min_length=1, max_length=100)
        description: Optional[str] = None
        price: Optional[float] = Field(None, gt=0)
        status: Optional[StatusEnum] = None
        tags: Optional[List[str]] = None
    
    class Item(ItemBase):
        """نموذج العنصر الكامل"""
        id: int
        created_at: datetime
        updated_at: datetime
        
        class Config:
            orm_mode = True
    
    class HealthResponse(BaseModel):
        """نموذج استجابة الصحة"""
        status: str
        timestamp: datetime
        version: str
    
    # قاعدة بيانات مؤقتة
    items_db: Dict[int, Dict[str, Any]] = {}
    next_id = 1
    
    # Middleware
    @app.middleware("http")
    async def log_requests(request, call_next):
        """تسجيل جميع الطلبات"""
        start_time = datetime.now()
        
        response = await call_next(request)
        
        duration = (datetime.now() - start_time).total_seconds()
        print(f"📝 {request.method} {request.url.path} - {response.status_code} ({duration:.3f}s)")
        
        return response
    
    # Exception Handlers
    @app.exception_handler(HTTPException)
    async def http_exception_handler(request, exc):
        """معالج الأخطاء HTTP"""
        return JSONResponse(
            status_code=exc.status_code,
            content={
                "error": {
                    "code": exc.status_code,
                    "message": exc.detail,
                    "timestamp": datetime.now().isoformat()
                }
            }
        )
    
    @app.exception_handler(Exception)
    async def general_exception_handler(request, exc):
        """معالج الأخطاء العام"""
        return JSONResponse(
            status_code=500,
            content={
                "error": {
                    "code": 500,
                    "message": "Internal server error",
                    "timestamp": datetime.now().isoformat()
                }
            }
        )
    
    # Routes
    @app.get("/", tags=["Root"])
    async def root():
        """الصفحة الرئيسية"""
        return {
            "message": "Welcome to the API",
            "version": "1.0.0",
            "docs": "/docs",
            "health": "/health"
        }
    
    @app.get("/health", response_model=HealthResponse, tags=["Health"])
    async def health_check():
        """فحص صحة الخادم"""
        return HealthResponse(
            status="healthy",
            timestamp=datetime.now(),
            version="1.0.0"
        )
    
    @app.get("/items", response_model=List[Item], tags=["Items"])
    async def list_items(
        skip: int = Query(0, ge=0),
        limit: int = Query(10, ge=1, le=100),
        status: Optional[StatusEnum] = None
    ):
        """قائمة جميع العناصر"""
        items = list(items_db.values())
        
        # تصفية حسب الحالة
        if status:
            items = [item for item in items if item["status"] == status]
        
        # تطبيق pagination
        return items[skip:skip + limit]
    
    @app.get("/items/{item_id}", response_model=Item, tags=["Items"])
    async def get_item(item_id: int):
        """الحصول على عنصر محدد"""
        if item_id not in items_db:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Item with id {item_id} not found"
            )
        
        return items_db[item_id]
    
    @app.post("/items", response_model=Item, status_code=status.HTTP_201_CREATED, tags=["Items"])
    async def create_item(item: ItemCreate):
        """إنشاء عنصر جديد"""
        global next_id
        
        now = datetime.now()
        new_item = {
            "id": next_id,
            "created_at": now,
            "updated_at": now,
            **item.dict()
        }
        
        items_db[next_id] = new_item
        next_id += 1
        
        return new_item
    
    @app.put("/items/{item_id}", response_model=Item, tags=["Items"])
    async def update_item(item_id: int, item_update: ItemUpdate):
        """تحديث عنصر"""
        if item_id not in items_db:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Item with id {item_id} not found"
            )
        
        stored_item = items_db[item_id]
        update_data = item_update.dict(exclude_unset=True)
        
        for field, value in update_data.items():
            stored_item[field] = value
        
        stored_item["updated_at"] = datetime.now()
        
        return stored_item
    
    @app.delete("/items/{item_id}", status_code=status.HTTP_204_NO_CONTENT, tags=["Items"])
    async def delete_item(item_id: int):
        """حذف عنصر"""
        if item_id not in items_db:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Item with id {item_id} not found"
            )
        
        del items_db[item_id]
        return None
    
    # Startup/Shutdown Events
    @app.on_event("startup")
    async def startup_event():
        """حدث بدء التشغيل"""
        print("🚀 API Server starting up...")
        print("📚 Documentation available at: http://localhost:8000/docs")
    
    @app.on_event("shutdown")
    async def shutdown_event():
        """حدث إيقاف التشغيل"""
        print("👋 API Server shutting down...")

def run_server(host: str = "0.0.0.0", port: int = 8000, reload: bool = False):
    """تشغيل الخادم"""
    if not FASTAPI_AVAILABLE:
        print("❌ FastAPI is not installed")
        return
    
    print(f"🌐 Starting server at http://{host}:{port}")
    uvicorn.run("api_server:app", host=host, port=port, reload=reload)

if __name__ == "__main__":
    run_server(reload=True)
```

---

### المرحلة 8️⃣: الاختبارات (Testing)

#### الأهداف:
- اختبارات وحدة شاملة
- اختبارات تكامل
- تغطية الكود

#### الأوامر:

```bash
# تثبيت أدوات الاختبار
pip install pytest pytest-cov pytest-asyncio pytest-mock

# تحديث requirements.txt
pip freeze | grep pytest >> requirements.txt

# إنشاء ملف إعدادات pytest
cat > pytest.ini << 'EOF'
[pytest]
testpaths = tests
python_files = test_*.py
python_classes = Test*
python_functions = test_*
addopts = 
    -v
    --strict-markers
    --cov=src
    --cov-report=html
    --cov-report=term-missing
markers =
    unit: Unit tests
    integration: Integration tests
    slow: Slow running tests
EOF

# Commit
git add pytest.ini requirements.txt
git commit -m "feat: add testing framework"
```

#### ملف Python: `tests/test_example.py`

```python
#!/usr/bin/env python3
"""
Example Tests - أمثلة على الاختبارات
"""

import pytest
from unittest.mock import Mock, patch, MagicMock
from typing import List

# Test Fixtures
@pytest.fixture
def sample_data():
    """بيانات اختبار نموذجية"""
    return {
        'id': 1,
        'name': 'Test Item',
        'value': 100
    }

@pytest.fixture
def sample_list():
    """قائمة اختبار نموذجية"""
    return [1, 2, 3, 4, 5]

# Unit Tests
class TestBasicOperations:
    """اختبارات العمليات الأساسية"""
    
    def test_addition(self):
        """اختبار الجمع"""
        assert 1 + 1 == 2
        assert 2 + 3 == 5
    
    def test_subtraction(self):
        """اختبار الطرح"""
        assert 5 - 3 == 2
        assert 10 - 7 == 3
    
    def test_multiplication(self):
        """اختبار الضرب"""
        assert 2 * 3 == 6
        assert 4 * 5 == 20
    
    def test_division(self):
        """اختبار القسمة"""
        assert 10 / 2 == 5
        assert 15 / 3 == 5
    
    def test_division_by_zero(self):
        """اختبار القسمة على صفر"""
        with pytest.raises(ZeroDivisionError):
            result = 10 / 0

class TestStringOperations:
    """اختبارات العمليات النصية"""
    
    def test_string_concatenation(self):
        """اختبار دمج النصوص"""
        assert "Hello" + " " + "World" == "Hello World"
    
    def test_string_formatting(self):
        """اختبار تنسيق النصوص"""
        name = "Ahmed"
        assert f"Hello, {name}!" == "Hello, Ahmed!"
    
    def test_string_methods(self):
        """اختبار دوال النصوص"""
        text = "Hello World"
        assert text.lower() == "hello world"
        assert text.upper() == "HELLO WORLD"
        assert text.replace("World", "Python") == "Hello Python"

class TestListOperations:
    """اختبارات العمليات على القوائم"""
    
    def test_list_append(self, sample_list):
        """اختبار إضافة عنصر"""
        sample_list.append(6)
        assert len(sample_list) == 6
        assert sample_list[-1] == 6
    
    def test_list_remove(self, sample_list):
        """اختبار حذف عنصر"""
        sample_list.remove(3)
        assert len(sample_list) == 4
        assert 3 not in sample_list
    
    def test_list_comprehension(self):
        """اختبار List Comprehension"""
        numbers = [1, 2, 3, 4, 5]
        squared = [n**2 for n in numbers]
        assert squared == [1, 4, 9, 16, 25]

class TestDictionaryOperations:
    """اختبارات العمليات على القواميس"""
    
    def test_dict_access(self, sample_data):
        """اختبار الوصول للقيم"""
        assert sample_data['id'] == 1
        assert sample_data['name'] == 'Test Item'
    
    def test_dict_update(self, sample_data):
        """اختبار تحديث القيم"""
        sample_data['value'] = 200
        assert sample_data['value'] == 200
    
    def test_dict_keys(self, sample_data):
        """اختبار المفاتيح"""
        keys = list(sample_data.keys())
        assert 'id' in keys
        assert 'name' in keys
        assert 'value' in keys

# Parameterized Tests
@pytest.mark.parametrize("input,expected", [
    (1, 2),
    (2, 4),
    (3, 6),
    (4, 8),
    (5, 10),
])
def test_double(input, expected):
    """اختبار مضاعفة الأرقام"""
    assert input * 2 == expected

@pytest.mark.parametrize("text,expected", [
    ("hello", "HELLO"),
    ("world", "WORLD"),
    ("Python", "PYTHON"),
])
def test_uppercase(text, expected):
    """اختبار تحويل لأحرف كبيرة"""
    assert text.upper() == expected

# Mocking Tests
class TestMocking:
    """اختبارات Mocking"""
    
    def test_mock_function(self):
        """اختبار Mock لدالة"""
        mock_func = Mock(return_value=42)
        result = mock_func()
        
        assert result == 42
        mock_func.assert_called_once()
    
    def test_mock_method(self):
        """اختبار Mock لـ Method"""
        mock_obj = Mock()
        mock_obj.method.return_value = "mocked"
        
        result = mock_obj.method("arg")
        
        assert result == "mocked"
        mock_obj.method.assert_called_once_with("arg")
    
    @patch('builtins.print')
    def test_patch_print(self, mock_print):
        """اختبار Patch لـ print"""
        print("Hello, World!")
        mock_print.assert_called_once_with("Hello, World!")

# Async Tests
@pytest.mark.asyncio
async def test_async_function():
    """اختبار دالة غير متزامنة"""
    async def async_add(a, b):
        return a + b
    
    result = await async_add(2, 3)
    assert result == 5

# Markers
@pytest.mark.unit
def test_unit_example():
    """مثال على اختبار وحدة"""
    assert True

@pytest.mark.integration
def test_integration_example():
    """مثال على اختبار تكامل"""
    assert True

@pytest.mark.slow
def test_slow_example():
    """مثال على اختبار بطيء"""
    import time
    time.sleep(0.1)
    assert True

# Skip and Xfail
@pytest.mark.skip(reason="Not implemented yet")
def test_skipped():
    """اختبار متخطى"""
    assert False

@pytest.mark.xfail(reason="Known bug")
def test_expected_failure():
    """اختبار متوقع فشله"""
    assert False

# Fixtures with Scope
@pytest.fixture(scope="module")
def module_fixture():
    """Fixture على مستوى الـ Module"""
    print("\n🔧 Module setup")
    yield "module data"
    print("\n🔧 Module teardown")

@pytest.fixture(scope="session")
def session_fixture():
    """Fixture على مستوى الـ Session"""
    print("\n🔧 Session setup")
    yield "session data"
    print("\n🔧 Session teardown")

def test_with_fixtures(module_fixture, session_fixture):
    """اختبار مع Fixtures"""
    assert module_fixture == "module data"
    assert session_fixture == "session data"

if __name__ == "__main__":
    pytest.main([__file__, "-v"])
```

---

### المرحلة 9️⃣: النشر والإنتاج (Deployment)

#### الأهداف:
- تجهيز التطبيق للإنتاج
- إنشاء Docker containers
- إعداد CI/CD

#### الأوامر:

```bash
# إنشاء Dockerfile
cat > Dockerfile << 'EOF'
FROM python:3.10-slim

WORKDIR /app

# تثبيت التبعيات
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# نسخ الكود
COPY . .

# المنفذ
EXPOSE 8000

# الأمر الافتراضي
CMD ["python", "main.py"]
EOF

# إنشاء docker-compose.yml
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8000:8000"
    environment:
      - ENV=production
      - DATABASE_URL=postgresql://user:password@db:5432/mydb
    depends_on:
      - db
      - redis
    volumes:
      - ./logs:/app/logs
    restart: unless-stopped
  
  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=mydb
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: unless-stopped
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
EOF

# إنشاء .dockerignore
cat > .dockerignore << 'EOF'
__pycache__
*.pyc
*.pyo
*.pyd
.Python
env/
venv/
.git
.gitignore
.vscode
.idea
*.log
*.db
*.sqlite3
.env
.DS_Store
Thumbs.db
EOF

# Commit
git add Dockerfile docker-compose.yml .dockerignore
git commit -m "feat: add Docker configuration"
```

#### ملف: `.github/workflows/ci.yml`

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    strategy:
      matrix:
        python-version: [3.8, 3.9, '3.10', '3.11']
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
    
    - name: Set up Python ${{ matrix.python-version }}
      uses: actions/setup-python@v4
      with:
        python-version: ${{ matrix.python-version }}
    
    - name: Cache pip packages
      uses: actions/cache@v3
      with:
        path: ~/.cache/pip
        key: ${{ runner.os }}-pip-${{ hashFiles('requirements.txt') }}
        restore-keys: |
          ${{ runner.os }}-pip-
    
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r requirements.txt
    
    - name: Run linting
      run: |
        pip install flake8 black
        flake8 src/ --count --select=E9,F63,F7,F82 --show-source --statistics
        black --check src/
    
    - name: Run tests
      run: |
        pytest tests/ -v --cov=src --cov-report=xml
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.xml
        fail_ci_if_error: true

  build:
    needs: test
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Log in to Docker Hub
      uses: docker/login-action@v2
      with:
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}
    
    - name: Build and push Docker image
      uses: docker/build-push-action@v4
      with:
        context: .
        push: true
        tags: |
          ${{ secrets.DOCKER_USERNAME }}/myapp:latest
          ${{ secrets.DOCKER_USERNAME }}/myapp:${{ github.sha }}
        cache-from: type=registry,ref=${{ secrets.DOCKER_USERNAME }}/myapp:latest
        cache-to: type=inline

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    
    steps:
    - name: Deploy to production
      run: |
        echo "🚀 Deploying to production..."
        # أضف أوامر النشر هنا
```

---

### المرحلة 🔟: المراقبة والصيانة (Monitoring & Maintenance)

#### الأهداف:
- مراقبة الأداء
- جمع المقاييس
- التنبيهات

#### ملف Python: `monitoring.py`

```python
#!/usr/bin/env python3
"""
Monitoring System - نظام المراقبة والمقاييس
"""

import time
import psutil
import platform
from typing import Dict, Any, Optional
from datetime import datetime
from dataclasses import dataclass, asdict
import json

@dataclass
class SystemMetrics:
    """مقاييس النظام"""
    timestamp: str
    cpu_percent: float
    memory_percent: float
    memory_used_mb: float
    memory_available_mb: float
    disk_percent: float
    disk_used_gb: float
    disk_free_gb: float
    network_sent_mb: float
    network_recv_mb: float

class SystemMonitor:
    """مراقب النظام"""
    
    def __init__(self):
        self.start_time = time.time()
        self.initial_net_io = psutil.net_io_counters()
    
    def get_system_info(self) -> Dict[str, Any]:
        """الحصول على معلومات النظام"""
        return {
            'platform': platform.system(),
            'platform_release': platform.release(),
            'platform_version': platform.version(),
            'architecture': platform.machine(),
            'hostname': platform.node(),
            'processor': platform.processor(),
            'python_version': platform.python_version(),
            'cpu_count': psutil.cpu_count(),
            'cpu_count_logical': psutil.cpu_count(logical=True),
        }
    
    def get_metrics(self) -> SystemMetrics:
        """الحصول على المقاييس الحالية"""
        # CPU
        cpu_percent = psutil.cpu_percent(interval=1)
        
        # Memory
        memory = psutil.virtual_memory()
        memory_percent = memory.percent
        memory_used_mb = memory.used / (1024 ** 2)
        memory_available_mb = memory.available / (1024 ** 2)
        
        # Disk
        disk = psutil.disk_usage('/')
        disk_percent = disk.percent
        disk_used_gb = disk.used / (1024 ** 3)
        disk_free_gb = disk.free / (1024 ** 3)
        
        # Network
        net_io = psutil.net_io_counters()
        network_sent_mb = (net_io.bytes_sent - self.initial_net_io.bytes_sent) / (1024 ** 2)
        network_recv_mb = (net_io.bytes_recv - self.initial_net_io.bytes_recv) / (1024 ** 2)
        
        return SystemMetrics(
            timestamp=datetime.now().isoformat(),
            cpu_percent=round(cpu_percent, 2),
            memory_percent=round(memory_percent, 2),
            memory_used_mb=round(memory_used_mb, 2),
            memory_available_mb=round(memory_available_mb, 2),
            disk_percent=round(disk_percent, 2),
            disk_used_gb=round(disk_used_gb, 2),
            disk_free_gb=round(disk_free_gb, 2),
            network_sent_mb=round(network_sent_mb, 2),
            network_recv_mb=round(network_recv_mb, 2)
        )
    
    def print_metrics(self, metrics: SystemMetrics) -> None:
        """طباعة المقاييس"""
        print("\n" + "="*60)
        print("System Metrics".center(60))
        print("="*60)
        print(f"📅 Timestamp: {metrics.timestamp}")
        print(f"\n💻 CPU: {metrics.cpu_percent}%")
        print(f"🧠 Memory: {metrics.memory_percent}% ({metrics.memory_used_mb:.0f} MB used)")
        print(f"💾 Disk: {metrics.disk_percent}% ({metrics.disk_used_gb:.1f} GB used)")
        print(f"🌐 Network: ⬆️ {metrics.network_sent_mb:.2f} MB | ⬇️ {metrics.network_recv_mb:.2f} MB")
        print("="*60)
    
    def check_alerts(self, metrics: SystemMetrics) -> list:
        """التحقق من التنبيهات"""
        alerts = []
        
        if metrics.cpu_percent > 80:
            alerts.append(f"⚠️ HIGH CPU: {metrics.cpu_percent}%")
        
        if metrics.memory_percent > 85:
            alerts.append(f"⚠️ HIGH MEMORY: {metrics.memory_percent}%")
        
        if metrics.disk_percent > 90:
            alerts.append(f"⚠️ HIGH DISK: {metrics.disk_percent}%")
        
        return alerts
    
    def save_metrics(self, metrics: SystemMetrics, filepath: str) -> None:
        """حفظ المقاييس إلى ملف"""
        try:
            with open(filepath, 'a') as f:
                f.write(json.dumps(asdict(metrics)) + '\n')
        except Exception as e:
            print(f"❌ Failed to save metrics: {e}")

def main():
    """الدالة الرئيسية"""
    monitor = SystemMonitor()
    
    # معلومات النظام
    print("\n🖥️  System Information:")
    info = monitor.get_system_info()
    for key, value in info.items():
        print(f"   {key}: {value}")
    
    # جمع المقاييس
    print("\n📊 Collecting metrics...")
    
    try:
        while True:
            metrics = monitor.get_metrics()
            monitor.print_metrics(metrics)
            
            # التحقق من التنبيهات
            alerts = monitor.check_alerts(metrics)
            if alerts:
                print("\n🚨 ALERTS:")
                for alert in alerts:
                    print(f"   {alert}")
            
            # حفظ المقاييس
            monitor.save_metrics(metrics, 'logs/metrics.jsonl')
            
            # انتظار
            time.sleep(5)
            
    except KeyboardInterrupt:
        print("\n\n👋 Monitoring stopped")

if __name__ == "__main__":
    main()
```

---

## 📚 المراجع والموارد

### الوثائق الرسمية
- **Python**: https://docs.python.org/
- **Git**: https://git-scm.com/doc
- **pip**: https://pip.pypa.io/
- **FastAPI**: https://fastapi.tiangolo.com/
- **SQLAlchemy**: https://docs.sqlalchemy.org/
- **pytest**: https://docs.pytest.org/

### أفضل الممارسات
1. **كتابة الكود النظيف**: اتبع PEP 8
2. **التوثيق**: وثق كل دالة وفئة
3. **الاختبارات**: اكتب اختبارات لكل feature
4. **Git Commits**: استخدم رسائل واضحة ومعبرة
5. **الأمان**: لا تحفظ بيانات حساسة في الكود

### الأدوات المفيدة
```bash
# Code Formatting
black src/
isort src/

# Linting
flake8 src/
pylint src/
mypy src/

# Security
bandit -r src/
safety check

# Documentation
pdoc --html src/
```

---

## ✅ قائمة المراجعة النهائية

- [ ] البيئة الافتراضية منشأة ومفعلة
- [ ] جميع التبعيات مثبتة
- [ ] هيكل المشروع منظم
- [ ] ملفات الإعدادات جاهزة
- [ ] نظام التسجيل يعمل
- [ ] معالج الأخطاء مطبق
- [ ] قاعدة البيانات متصلة
- [ ] API يعمل بشكل صحيح
- [ ] الاختبارات تمر بنجاح
- [ ] Docker containers جاهزة
- [ ] CI/CD pipeline معدة
- [ ] نظام المراقبة يعمل
- [ ] الوثائق مكتملة
- [ ] الكود موثق
- [ ] Git history نظيف

---

## 🎉 الخاتمة

هذا البرومبت الشامل يغطي جميع مراحل تطوير المشروع من البداية حتى النشر والمراقبة. استخدمه كمرجع ودليل لبناء تطبيقات احترافية وقابلة للتوسع.

**للمساعدة أو الاستفسارات:**
- افتح Issue على GitHub
- راجع الوثائق
- انضم إلى المجتمع

**Good luck! 🚀**
