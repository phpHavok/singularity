/* 
 * Copyright (c) 2015-2016, Gregory M. Kurtzer. All rights reserved.
 * 
 * “Singularity” Copyright (c) 2016, The Regents of the University of California,
 * through Lawrence Berkeley National Laboratory (subject to receipt of any
 * required approvals from the U.S. Dept. of Energy).  All rights reserved.
 * 
 * This software is licensed under a customized 3-clause BSD license.  Please
 * consult LICENSE file distributed with the sources of this project regarding
 * your rights to use or distribute this software.
 * 
 * NOTICE.  This Software was developed under funding from the U.S. Department of
 * Energy and the U.S. Government consequently retains certain rights. As such,
 * the U.S. Government has been granted for itself and others acting on its
 * behalf a paid-up, nonexclusive, irrevocable, worldwide license in the Software
 * to reproduce, distribute copies to the public, prepare derivative works, and
 * perform publicly and display publicly, and to permit other to do so. 
 * 
*/

#include <errno.h>
#include <fcntl.h>
#include <stdio.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <limits.h>
#include <unistd.h>
#include <stdlib.h>
#include <grp.h>
#include <pwd.h>


#include "util/file.h"
#include "util/util.h"
#include "message.h"
#include "lib/privilege.h"
#include "lib/sessiondir.h"
#include "lib/rootfs/rootfs.h"
#include "lib/file/file.h"


int singularity_file_passwd(void) {
    FILE *file_fp;
    char *source_file;
    char *tmp_file;
    uid_t uid = singularity_priv_getuid();
    struct passwd *pwent = getpwuid(uid);
    char *containerdir = singularity_rootfs_dir();
    char *sessiondir = singularity_sessiondir_get();

    message(DEBUG, "Called singularity_file_passwd_create()\n");

    if ( uid == 0 ) {
        message(VERBOSE, "Not updating passwd file, running as root!\n");
        return(0);
    }

    if ( containerdir == NULL ) {
        message(ERROR, "Failed to obtain container directory\n");
        ABORT(255);
    }

    if ( sessiondir == NULL ) {
        message(ERROR, "Failed to obtain session directory\n");
        ABORT(255);
    }

    source_file = joinpath(containerdir, "/etc/passwd");
    tmp_file = joinpath(sessiondir, "/passwd");

    message(VERBOSE2, "Checking for template passwd file: %s\n", source_file);
    if ( is_file(source_file) < 0 ) {
        message(VERBOSE, "Passwd file does not exist in container, not updating\n");
        return(0);
    }

    message(VERBOSE2, "Creating template of /etc/passwd\n");
    if ( ( copy_file(source_file, tmp_file) ) < 0 ) {
        message(ERROR, "Failed copying template passwd file to sessiondir: %s\n", strerror(errno));
        ABORT(255);
    }

    message(DEBUG, "Opening the template passwd file: %s\n", tmp_file);
    if ( ( file_fp = fopen(tmp_file, "a") ) == NULL ) { // Flawfinder: ignore
        message(ERROR, "Could not open template passwd file %s: %s\n", tmp_file, strerror(errno));
        ABORT(255);
    }

    message(VERBOSE, "Creating template passwd file and appending user data\n");
    if ( ( file_fp = fopen(tmp_file, "a") ) == NULL ) { // Flawfinder: ignore
        message(ERROR, "Could not open template passwd file %s: %s\n", tmp_file, strerror(errno));
        ABORT(255);
    }
    fprintf(file_fp, "\n%s:x:%d:%d:%s:%s:%s\n", pwent->pw_name, pwent->pw_uid, pwent->pw_gid, pwent->pw_gecos, pwent->pw_dir, pwent->pw_shell);
    fclose(file_fp);


    container_file_bind("passwd", "/etc/passwd");

    return(0);
}
