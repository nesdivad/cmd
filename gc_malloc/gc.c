#include <stdio.h>

/*
Malloc implementations maintains a linked-list of free blocks of memory that can be partitioned and given out as needed.
*/

typedef struct header
{
    unsigned int size;
    struct header *next;
} header_t;

static header_t base_block;                     // Zero-sized block to get us started
static header_t *free_pointer = &base_block;    // Points to first free block of memory
static header_t *used_pointer;                  // Points to first used block of memory
static uintptr_t *stack_bottom;

/*
Scan the free list and look for a place to put the block.
We're looking for any block that the to-be-freed block might have been partitioned from.
*/
static void 
add_to_free_list(header_t *block_pointer) {
    header_t *pointer;

    for (pointer = free_pointer; !(block_pointer > pointer && block_pointer < pointer->next); pointer = pointer->next) {
        if (pointer >= pointer->next && (block_pointer > pointer) || block_pointer < pointer->next)
            break;
    }

    if (block_pointer + block_pointer->size == pointer->next) {
        block_pointer->size += pointer->next->size;
        block_pointer->next = pointer->next->next;
    } else {
        block_pointer->next = pointer->next;
    }

    if (pointer + pointer->size == block_pointer) {
        pointer->size += block_pointer->size;
        pointer->next = block_pointer->next;
    } else {
        pointer->next = block_pointer;
    }

    free_pointer = pointer;
}

// We allocate blocks in page sized chunks
#define MIN_ALLOC_SIZE 4096

/*
Request more memory from the kernel
*/
static header_t *morecore(size_t num_units) {
    void *void_pointer;
    header_t *u_pointer;

    if (num_units > MIN_ALLOC_SIZE) {
        num_units = MIN_ALLOC_SIZE / sizeof(header_t);
    }

    // -1 means failed
    if ((void_pointer = sbrk(num_units * sizeof(header_t))) == (void *) -1) {
        return NULL;
    }

    u_pointer = (header_t *) void_pointer;
    u_pointer->size = num_units;
    add_to_free_list(u_pointer);
    return free_pointer;
}

void *GC_malloc(size_t alloc_size) {
    size_t num_units;
    header_t *pointer, *previous_pointer;

    num_units = (alloc_size + sizeof(header_t) - 1) / sizeof(header_t) + 1;
    previous_pointer = free_pointer;

    for (pointer = previous_pointer->next;; previous_pointer = pointer, pointer = pointer->next) {
        if (pointer->size >= num_units) {
            if (pointer->size == num_units) {
                previous_pointer->next = pointer->next;
            } else { 
                // split the block into suitable size
                pointer->size -= num_units;
                pointer += pointer->size;
                pointer->size = num_units;
            }

            free_pointer = previous_pointer;

            if (used_pointer == NULL) {
                used_pointer = pointer->next = pointer;
            } else {
                pointer->next = used_pointer->next;
                used_pointer->next = pointer;
            }

            return (void *) (pointer + 1);
        }
        if (pointer == free_pointer) {
            pointer = morecore(num_units);
            if (pointer == NULL) {
                return NULL;
            }
        }
    }
}

#define UNTAG(p) (((uintptr_t) (p)) & 0xfffffffc)

/*
Scan a region of memory and mark any items in the used list appropriately.
Both arguments should be word aligned.
*/
static void scan_region(uintptr_t *shared_pointer, uintptr_t *end) {
    header_t *block_pointer;

    for (; shared_pointer < end; shared_pointer++) {
        uintptr_t v = *shared_pointer;
        block_pointer = used_pointer;
        do {
            if (block_pointer + 1 <= v && block_pointer + 1 + block_pointer->size > v) {
                block_pointer->next = ((uintptr_t) block_pointer->next) | 1;
                break;
            }
        } while ((block_pointer = UNTAG(block_pointer->next)) != used_pointer);
    }
}

/*
Scan the marked blocks for references to other unmarked blocks
*/
static void scan_heap(void) {
    uintptr_t *void_pointer;
    header_t *block_pointer, *u_pointer;

    for (block_pointer = UNTAG(used_pointer->next); block_pointer != used_pointer; block_pointer = UNTAG(block_pointer->next)) {
        if (!((uintptr_t) block_pointer->next) & 1)
            continue;

        for (void_pointer = (uintptr_t *)(block_pointer + 1); void_pointer < (block_pointer + block_pointer->size + 1); void_pointer++) {
            uintptr_t v = *void_pointer;
            u_pointer = UNTAG(block_pointer->next);
            do {
                if (u_pointer != block_pointer && u_pointer + 1 <= v && u_pointer + 1 + u_pointer->size > v) {
                    u_pointer->next = ((uintptr_t) u_pointer->next) | 1;
                    break;
                }
            } while ((u_pointer = UNTAG(u_pointer->next)) != block_pointer);
        }
    }
}

void GC_init(void) {
    static int initted;
    FILE *statfp;

    if (initted) return;

    initted = 1;

    statfp = fopen("/proc/self/stat", "r");
    assert(statfp != NULL);
    fscanf(statfp, 
            "%*d %*s %*c %*d %*d %*d %*d %*d %*u "
            "%*lu %*lu %*lu %*lu %*lu %*lu %*ld %*ld "
            "%*ld %*ld %*ld %*ld %*llu %*lu %*ld "
            "%*lu %*lu %*lu %lu", &stack_bottom);
    fclose(statfp);

    used_pointer = NULL;
    base_block.next = free_pointer = &base_block;
    base_block.size = 0;
}

void GC_collect(void) {
    header_t *pointer, *previous_pointer, *t_pointer;
    uintptr_t stack_top;
    extern char end, etext; // provided by the linker

    if (used_pointer == NULL) return;

    scan_region(&etext, &end);

    // Scan the volatile
    asm volatile ("movl %%ebp, %0" : "=r" (stack_top));
    scan_region(stack_top, stack_bottom);

    // mark from the heap
    scan_heap();

    for (previous_pointer = used_pointer, pointer = UNTAG(used_pointer->next);; previous_pointer = pointer, pointer = UNTAG(pointer->next)) {
    next_chunk:
        if (!((unsigned int)pointer->next & 1)) {
            t_pointer = pointer;
            pointer = UNTAG(pointer->next);
            add_to_free_list(t_pointer);

            if (used_pointer == t_pointer) {
                used_pointer = NULL;
                break;
            }

            previous_pointer->next = (uintptr_t)pointer | ((uintptr_t)previous_pointer->next & 1);
            goto next_chunk;
        }

        pointer->next = ((uintptr_t) pointer->next & ~1);
        if (pointer == used_pointer) break;
    }
}