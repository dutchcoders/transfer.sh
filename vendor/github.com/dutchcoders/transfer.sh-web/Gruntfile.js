'use strict';

// # Globbing
// for performance reasons we're only matching one level down:
// 'test/spec/{,*/}*.js'
// use this if you want to match all subfolders:
// 'test/spec/**/*.js'

module.exports = function (grunt) {
    // load all grunt tasks
    require('load-grunt-tasks')(grunt);
    // show elapsed time at the end
    require('time-grunt')(grunt);

    grunt.loadNpmTasks('grunt-npm-command');

    // configurable paths
    var yeomanConfig = {
        app: require('./bower.json').appPath || 'src',
        dist: 'dist/'
    };

    grunt.initConfig({
        yeoman: yeomanConfig,
        watch: {
            less: {
                files: ['<%= yeoman.app %>/styles/{,*/}*.less'],
                tasks: ['less']
            },
            gruntfile: {
                files: ['Gruntfile.js']
            },
            includes: {
                files: ['<%= yeoman.app %>/*.html', '.tmp/*.html'],
                tasks: ['includes:server']
            },
            livereload: {
                options: {
                    livereload: '<%= connect.options.livereload %>'
                },
                files: [
                    '<%= yeoman.app %>/*.html',
                    '{.tmp,<%= yeoman.app %>}/styles/{,*/}*.css',
                    '{.tmp,<%= yeoman.app %>}/scripts/{,*/}*.js',
                    '<%= yeoman.app %>/images/{,*/}*.{png,jpg,jpeg,gif,webp,svg}'
                ],
                tasks: ['includes:server']
            }
        },
        connect: {
            options: {
                port: 9000,
                // change this to '0.0.0.0' to access the server from outside
                hostname: 'localhost',
                livereload: 35729
            },
            livereload: {
                options: {
                    open: true,
                    base: [
                        '.tmp',
                        '<%= yeoman.app %>'
                    ]
                }
            },
            test: {
                options: {
                    port: 9001,
                    base: [
                        '.tmp',
                        'test',
                        '<%= yeoman.app %>'
                    ]
                }
            },
            dist: {
                options: {
                    base: '<%= yeoman.dist %>'
                }
            }
        },
        clean: {
            dist: {
                files: [{
                    dot: true,
                    src: [
                        '.tmp',
                        '<%= yeoman.dist %>/*',
                        '!<%= yeoman.dist %>/.git*'
                    ]
                }]
            },
            server: '.tmp'
        },
        jshint: {
            options: {
                jshintrc: '.jshintrc',
                reporter: require('jshint-stylish')
            },
            all: [
                'Gruntfile.js',
                '<%= yeoman.app %>/scripts/{,*/}*.js',
                '!<%= yeoman.app %>/scripts/vendor/*',
                'test/spec/{,*/}*.js'
            ]
        },


        less: {
            dist: {
                files: {
                    '<%= yeoman.app %>/styles/main.css': ['<%= yeoman.app %>/styles/main.less']
                },
                options: {
                    sourceMap: true,
                    sourceMapFilename: '<%= yeoman.app %>/styles/main.css.map',
                    sourceMapBasepath: '<%= yeoman.app %>/',
                    sourceMapRootpath: '/'
                }
            }
        },

        includes: {
            build: {
                cwd: '<%= yeoman.app %>',
                src: ['*.html', 'includes/*.html'],
                dest: '<%= yeoman.dist %>',
                options: {
                    flatten: true,
                    banner: ''
                }
            },
            server: {
                cwd: '<%= yeoman.app %>',
                src: ['*.html', 'includes/*.html'],
                dest: '.tmp/',
                options: {
                    flatten: true,
                    banner: ''
                }
            }
        },
        // not used since Uglify task does concat,
        // but still available if needed
        /*concat: {
      dist: {}
    },*/
        // not enabled since usemin task does concat and uglify
        // check index.html to edit your build targets
        // enable this task if you prefer defining your build targets here
        /*uglify: {
      dist: {}
    },*/
        rev: {
            dist: {
                files: {
                    src: [
                        '<%= yeoman.dist %>/scripts/{,*/}*.js',
                        '<%= yeoman.dist %>/styles/{,*/}*.css',
                        '<%= yeoman.dist %>/images/{,*/}*.{png,jpg,jpeg,gif,webp,svg}',
                        '<%= yeoman.dist %>/fonts/{,*/}*.*'
                    ]
                }
            }
        },
        useminPrepare: {
            html: '<%= yeoman.app %>/*.html',
            options: {
                dest: '<%= yeoman.dist %>'
            }
        },
        usemin: {
            html: ['<%= yeoman.dist %>/{,*/}*.html'],
            css: ['<%= yeoman.dist %>/styles/{,*/}*.css'],
            options: {
                dirs: ['<%= yeoman.dist %>']
            }
        },
        imagemin: {
            dist: {
                files: [{
                    expand: true,
                    cwd: '<%= yeoman.app %>/images',
                    src: '{,*/}*.{png,jpg,jpeg}',
                    dest: '<%= yeoman.dist %>/images'
                }]
            }
        },

        cssmin: {
            dist: {
                files: {
                    '<%= yeoman.dist %>/styles/main.css': [
                        '.tmp/styles/{,*/}*.css',
                        '<%= yeoman.app %>/styles/{,*/}*.css'
                    ]
                }
            }
        },
        htmlmin: {
            dist: {
                options: {
                    /*removeCommentsFromCDATA: true,
          // https://github.com/yeoman/grunt-usemin/issues/44
          //collapseWhitespace: true,
          collapseBooleanAttributes: true,
          removeAttributeQuotes: true,
          removeRedundantAttributes: true,
          useShortDoctype: true,
          removeEmptyAttributes: true,
          removeOptionalTags: true*/
                },
                files: [{
                    expand: true,
                    cwd: '<%= yeoman.app %>',
                    src: '*.html',
                    dest: '<%= yeoman.dist %>'
                }]
            }
        },
        'npm-command': {
            'videojs-install': {
                options: {
                    cwd: '<%= yeoman.app %>/bower_components/videojs/'
                }
            },
            'videojs-build': {
                options: {
                    cmd: 'run-script',
                    args: ['build'],
                    cwd: '<%= yeoman.app %>/bower_components/videojs/'
                }
            }
        }
        ,copy: {
            dist: {
                files: [{
                    expand: true,
                    dot: true,
                    cwd: '<%= yeoman.app %>',
                    dest: '<%= yeoman.dist %>',
                    src: [
                        '*.{ico,png,txt}',
                        'fonts/{,*/}*.*',
                        '.htaccess',
                        'index.txt',
                        '404.txt',
                        'images/{,*/}*.{webp,gif,svg}'
                    ]
                }]
            },
            server: {
                files: [{
                    expand: true,
                    dot: true,
                    cwd: '<%= yeoman.app %>/bower_components/font-awesome/fonts/',
                    dest: '<%= yeoman.app %>/fonts/font-awesome',
                    src: ['*']
                }, {
                    expand: true,
                    dot: true,
                    cwd: '<%= yeoman.app %>/bower_components/bootstrap/dist/fonts/',
                    dest: '<%= yeoman.app %>/fonts/glyphicons',
                    src: ['*']
                }]
            }
        },
        concurrent: {
            dist: [
                'less',
                'imagemin',
                'htmlmin'
            ]
        }
    });

    grunt.loadNpmTasks('grunt-npm-command');

    grunt.registerTask('serve', function (target) {
        if (target === 'dist') {
            return grunt.task.run(['build', 'connect:dist:keepalive']);
        }

        grunt.task.run([
            'clean:server',
            'less',
            'includes:server',
            'copy:server',
            'connect:livereload',
            'watch'
        ]);
    });

    grunt.registerTask('server', function () {
        grunt.log.warn('The `server` task has been deprecated. Use `grunt serve` to start a server.');
        grunt.task.run(['serve']);
    });

    grunt.registerTask('test', [
        'clean:server',
        'less',
        'copy:server',
        'connect:test',
    ]);

    grunt.registerTask('build', [
        'clean:dist',
        'npm-command',
        'copy:server',
        'useminPrepare',
        'concurrent',
        'cssmin',
        'concat',
        'includes:build',
        'uglify',
        'copy',
        'usemin',

    ]);

    grunt.registerTask('default', [
        'jshint',
        'test',
        'build'
    ]);
};
