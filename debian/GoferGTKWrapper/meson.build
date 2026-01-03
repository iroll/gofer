project('gofer', 'vala', 'c',
  version: '0.9.0',
  meson_version: '>= 0.59.0',
  default_options: [
    'warning_level=2',
  ]
)

# Dependencies
gtk4_dep = dependency('gtk4')
adwaita_dep = dependency('libadwaita-1', version: '>= 1.0')

# Get installation directories
prefix = get_option('prefix')
bindir = prefix / get_option('bindir')
datadir = prefix / get_option('datadir')
applicationsdir = datadir / 'applications'

# Build gofer Go binary
#go = find_program('go')
#gofer_binary = custom_target('gofer',
#  output: 'gofer',
#  command: [go, 'build', '-o', '@OUTPUT@'],
#  build_by_default: true,
#  install: true,
#  install_dir: bindir,
#  console: true,
#  env: {'CGO_ENABLED': '0'},
#  depend_files: files('../../gofer.go', '../../gofer_engine.go', '../../go.mod')
#)

# Configure desktop file
desktop_conf = configuration_data()
desktop_conf.set('BINDIR', bindir)

desktop_file = configure_file(
  input: 'data/gofer.desktop.in',
  output: 'gofer.desktop',
  configuration: desktop_conf,
  install: true,
  install_dir: applicationsdir
)

# Build the wrapper
executable('gofer-wrapper',
  'GoferGTKWrapper.vala',
  dependencies: [
    gtk4_dep,
    adwaita_dep,
  ],
  install: true,
  install_dir: bindir
)

