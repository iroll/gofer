project('gofer', 'vala', 'c',
  version: '0.9.2',
  meson_version: '>= 0.59.0',
  default_options: [
    'warning_level=2',
  ]
)

# Canonical application ID (DO NOT CHANGE)
app_id = 'com.iroll.gofer'

# Dependencies
gtk4_dep = dependency('gtk4')
adwaita_dep = dependency('libadwaita-1', version: '>= 1.0')

# Get installation directories
prefix = get_option('prefix')
bindir = prefix / get_option('bindir')
datadir = prefix / get_option('datadir')
applicationsdir = datadir / 'applications'

# Configure desktop file
desktop_conf = configuration_data()
desktop_conf.set('BINDIR', bindir)
desktop_conf.set('APP_ID', app_id)

desktop_file = configure_file(
  input: 'data/' + app_id + '.desktop.in',
  output: app_id + '.desktop',
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

