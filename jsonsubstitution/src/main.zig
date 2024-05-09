const std = @import("std");
pub const Value = std.json.Value;

fn readJsonFile(allocator: std.mem.Allocator, path: []u8) !std.json.Parsed(Value) {
    const file = try std.fs.cwd().readFileAlloc(allocator, path, 512);
    defer allocator.free(file);
    return std.json.parseFromSlice(Value, allocator, file, .{ .allocate = .alloc_always, .ignore_unknown_fields = true });
}

fn writeToFile(value: std.StringArrayHashMapUnmanaged(Value), name: []u8) !void {
    var out = std.ArrayList(u8).init(std.heap.page_allocator);
    var writeStream = std.json.writeStream(out.writer(), .{ .whitespace = .indent_4 });
    defer writeStream.deinit();

    try writeStream.beginObject();
    var it = value.iterator();
    while (it.next()) |entry| {
        std.debug.print("\nKey: {s}, Value: {?}\n", .{ entry.key_ptr.*, entry.value_ptr.* });
        try writeStream.objectField(entry.key_ptr.*);
        try writeStream.write(entry.value_ptr.*);
    }
    try writeStream.endObject();

    var file = std.fs.cwd().createFile(name, .{ .read = true, .exclusive = true }) catch |err| switch (err) {
        error.PathAlreadyExists => {
            std.debug.print("\nFile already exists: {s}", .{name});
            return;
        },
        else => {
            std.debug.print("{?}", .{err});
            return;
        },
    };
    try file.writeAll(out.items);
}

pub fn compareAndReplace(original: Value, into: Value, allocator: std.mem.Allocator) !std.StringArrayHashMapUnmanaged(Value) {
    var originalHMap = try std.json.ArrayHashMap(Value).jsonParseFromValue(allocator, original, .{ .allocate = .alloc_always });
    var intoHMap = try std.json.ArrayHashMap(Value).jsonParseFromValue(allocator, into, .{ .allocate = .alloc_always });

    defer originalHMap.deinit(allocator);
    defer intoHMap.deinit(allocator);

    var iter2 = intoHMap.map.iterator();
    while (iter2.next()) |entry| {
        _ = try originalHMap.map.put(allocator, entry.key_ptr.*, entry.value_ptr.*);
    }

    var iter = originalHMap.map.iterator();
    while (iter.next()) |entry| {
        std.debug.print("\nKey/Value: {s}, {?}", .{ entry.key_ptr.*, entry.value_ptr.* });
    }

    return originalHMap.map;
}

pub fn main() !void {
    const allocator = std.heap.page_allocator;

    const argv = std.process.argsAlloc(allocator) catch |err| {
        std.debug.print("Alloc failed: {?}", .{err});
        return;
    };
    defer std.process.argsFree(allocator, argv);
    const filepath1 = argv[1];
    const filepath2 = argv[2];
    const filename = argv[3];

    const file1 = try readJsonFile(allocator, filepath1);
    defer file1.deinit();

    const file2 = try readJsonFile(allocator, filepath2);
    defer file2.deinit();

    const hashmap = compareAndReplace(file1.value, file2.value, allocator) catch |err| {
        std.debug.print("{?}", .{err});
        return;
    };
    try writeToFile(hashmap, filename);
}
