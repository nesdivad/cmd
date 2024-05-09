const std = @import("std");
pub const Value = std.json.Value;

fn readJsonFile(allocator: std.mem.Allocator, path: []u8) !std.json.Parsed(Value) {
    const file = try std.fs.cwd().readFileAlloc(allocator, path, 512);
    defer allocator.free(file);
    return std.json.parseFromSlice(Value, allocator, file, .{ .allocate = .alloc_always, .ignore_unknown_fields = true });
}

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    const allocator = gpa.allocator();

    const argv = std.process.argsAlloc(allocator) catch |err| {
        std.debug.print("Alloc failed: {?}", .{err});
        return;
    };
    defer std.process.argsFree(allocator, argv);
    const filepath1 = argv[1];
    const filepath2 = argv[2];

    const file1 = try readJsonFile(allocator, filepath1);
    defer file1.deinit();

    const file2 = try readJsonFile(allocator, filepath2);
    defer file2.deinit();

    std.debug.print("file1: {?}\n", .{file1.value});
    std.debug.print("file2: {?}\n", .{file2.value});
}
