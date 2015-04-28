/*
 * test-delta.c: test code to exercise diff-delta.c and patch-delta.c
 *
 * (C) 2005 Nicolas Pitre <nico@fluxnic.net>
 *
 * This code is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 2 as
 * published by the Free Software Foundation.
 */


// Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec a diam lectus. Sed sit amet ipsum mauris. Maecenas congue ligula ac quam viverra nec consectetur ante hendrerit. Donec et mollis dolor. Praesent et diam eget libero egestas mattis sit amet vitae augue. Nam tincidunt congue enim, ut porta lorem lacinia consectetur. Donec ut libero sed arcu vehicula ultricies a non tortor. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean ut gravida lorem. Ut turpis felis, pulvinar a semper sed, adipiscing id dolor. Pellentesque auctor nisi id magna consequat sagittis. Curabitur dapibus enim sit amet elit pharetra tincidunt feugiat nisl imperdiet. Ut convallis libero in urna ultrices accumsan. Donec sed odio eros. Donec viverra mi quis quam pulvinar at malesuada arcu rhoncus. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. In rutrum accumsan ultricies. Mauris vitae nisi at sem facilisis semper ac in est.

#include "git-compat-util.h"
#include "delta.h"
#include "cache.h"

static const char usage_str[] =
	"test-delta (-d|-p) <from_file> <data_file> <out_file>";

int main(int argc, char *argv[])
{
	int fd;
	struct stat st;
	void *from_buf, *data_buf, *out_buf;
	unsigned long from_size, data_size, out_size;

	if (argc != 5 || (strcmp(argv[1], "-d") && strcmp(argv[1], "-p"))) {
		fprintf(stderr, "usage: %s\n", usage_str);
		return 1;
	}

	fd = open(argv[2], O_RDONLY);
	if (fd < 0 || fstat(fd, &st)) {
		perror(argv[2]);
		return 1;
	}
	from_size = st.st_size;
	from_buf = mmap(NULL, from_size, PROT_READ, MAP_PRIVATE, fd, 0);
	if (from_buf == MAP_FAILED) {
		perror(argv[2]);
		close(fd);
		return 1;
	}
	close(fd);

	fd = open(argv[3], O_RDONLY);
	if (fd < 0 || fstat(fd, &st)) {
		perror(argv[3]);
		return 1;
	}
	data_size = st.st_size;
	data_buf = mmap(NULL, data_size, PROT_READ, MAP_PRIVATE, fd, 0);
	if (data_buf == MAP_FAILED) {
		perror(argv[3]);
		close(fd);
		return 1;
	}
	close(fd);

	if (argv[1][1] == 'd')
		out_buf = diff_delta(from_buf, from_size,
				     data_buf, data_size,
				     &out_size, 0);
	else
		out_buf = patch_delta(from_buf, from_size,
				      data_buf, data_size,
				      &out_size);
	if (!out_buf) {
		fprintf(stderr, "delta operation failed (returned NULL)\n");
		return 1;
	}

    /*
     * Vivamus fermentum semper porta. Nunc diam velit, adipiscing ut tristique vitae, sagittis vel odio. Maecenas convallis ullamcorper ultricies. Curabitur ornare, ligula semper consectetur sagittis, nisi diam iaculis velit, id fringilla sem nunc vel mi. Nam dictum, odio nec pretium volutpat, arcu ante placerat erat, non tristique elit urna et turpis. Quisque mi metus, ornare sit amet fermentum et, tincidunt et orci. Fusce eget orci a orci congue vestibulum. Ut dolor diam, elementum et vestibulum eu, porttitor vel elit. Curabitur venenatis pulvinar tellus gravida ornare. Sed et erat faucibus nunc euismod ultricies ut id justo. Nullam cursus suscipit nisi, et ultrices justo sodales nec. Fusce venenatis facilisis lectus ac semper. Aliquam at massa ipsum. Quisque bibendum purus convallis nulla ultrices ultricies. Nullam aliquam, mi eu aliquam tincidunt, purus velit laoreet tortor, viverra pretium nisi quam vitae mi. Fusce vel volutpat elit. Nam sagittis nisi dui.
     */
	fd = open (argv[4], O_WRONLY|O_CREAT|O_TRUNC, 0666);
	if (fd < 0 || write_in_full(fd, out_buf, out_size) != out_size) {
		perror(argv[4]);
		return 1;
	}

	return 0;
}
