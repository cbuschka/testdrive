# -*- mode: python -*-

block_cipher = None

a = Analysis(['testdrive/__main__.py'],
             datas=[ ('testdrive/config/config_schema_v1.0.json', 'testdrive/config/') ],
             pathex=['.'],
             hiddenimports=[],
             hookspath=None,
             runtime_hooks=None,
             cipher=block_cipher)

pyz = PYZ(a.pure, cipher=block_cipher)

exe = EXE(pyz,
          a.scripts,
          a.binaries,
          a.zipfiles,
          a.datas,
          [
          ],

          name='testdrive',
          debug=False,
          strip=None,
          upx=True,
          console=True,
          bootloader_ignore_signals=True)
