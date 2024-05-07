const std = @import("std");
const allocator = std.heap.page_allocator;

pub fn openfile(path: []u8) std.fs.File.OpenError!std.fs.File {
    return std.fs.cwd().openFile(path, .{ .mode = .read_only });
}

pub fn dump_hex(data: []u8) void {
    const linelen = 16;
    const chunklen = 2;

    var offset: u8 = 0;

    var lit = std.mem.window(u8, data, linelen, linelen);
    while (lit.next()) |line| {
        std.debug.print("{x:0>8}: ", .{offset});

        var cit = std.mem.window(u8, line, chunklen, chunklen);
        while (cit.next()) |chunk| {
            var tempchunk: [chunklen]u8 = undefined;
            std.mem.copyForwards(u8, &tempchunk, chunk);

            const hexchunk = std.fmt.bytesToHex(tempchunk, .lower);
            std.debug.print("{s} ", .{hexchunk});
        }

        std.debug.print(" ", .{});
        std.debug.print("\n", .{});

        offset += linelen;
    }
}

pub fn main() !void {
    const argv = std.process.argsAlloc(allocator) catch |err| {
        std.debug.print("Alloc failed {?}\n", .{err});
        return;
    };
    defer std.process.argsFree(allocator, argv);

    const filepath = argv[1];
    const file = openfile(filepath) catch |err| switch (err) {
        error.FileNotFound => {
            std.debug.print("Could not find file: {s}", .{filepath});
            return;
        },
        else => {
            std.debug.print("{?}\n", .{err});
            return;
        },
    };

    const fStats = try file.stat();

    const buf = try allocator.alloc(u8, fStats.size);
    _ = try file.read(buf);
    dump_hex(buf);
}
