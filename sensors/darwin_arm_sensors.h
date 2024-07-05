// This code is originally from https://github.com/freedomtan/sensors/blob/master/sensors/sensors.m
// Here is the original code's license

// BSD 3-Clause License

// Copyright (c) 2016-2018, "freedom" Koan-Sin Tan
// All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:

// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.

// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.

// * Neither the name of the copyright holder nor the names of its
//   contributors may be used to endorse or promote products derived from
//   this software without specific prior written permission.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
#import <Foundation/Foundation.h>
#import <IOKit/hidsystem/IOHIDEventSystemClient.h>
#include <unistd.h>

typedef struct __IOHIDEvent         *IOHIDEventRef;
typedef struct __IOHIDServiceClient *IOHIDServiceClientRef;
typedef double                      IOHIDFloat;

IOHIDEventSystemClientRef IOHIDEventSystemClientCreate(CFAllocatorRef allocator);

int IOHIDEventSystemClientSetMatching(IOHIDEventSystemClientRef client, CFDictionaryRef match);

IOHIDEventRef IOHIDServiceClientCopyEvent(IOHIDServiceClientRef, int64_t, int32_t, int64_t);

CFStringRef IOHIDServiceClientCopyProperty(IOHIDServiceClientRef service, CFStringRef property);

IOHIDFloat IOHIDEventGetFloatValue(IOHIDEventRef event, int32_t field);

NSDictionary *matching(int page, int usage) {
    NSDictionary *dict = @{
        @"PrimaryUsagePage" : [NSNumber numberWithInt:page],
        @"PrimaryUsage" : [NSNumber numberWithInt:usage],
    };

    return dict;
}

NSArray *getProductNames(NSDictionary *sensors) {
    IOHIDEventSystemClientRef system = IOHIDEventSystemClientCreate(kCFAllocatorDefault);

    IOHIDEventSystemClientSetMatching(system, (__bridge CFDictionaryRef)sensors);
    NSArray *matchingsrvs = (__bridge NSArray *)IOHIDEventSystemClientCopyServices(system);

    long            count = [matchingsrvs count];
    NSMutableArray  *array = [[NSMutableArray alloc] init];

    for (int i = 0; i < count; i++) {
        IOHIDServiceClientRef   sc = (IOHIDServiceClientRef)matchingsrvs[i];
        NSString                *name = (NSString *)IOHIDServiceClientCopyProperty(sc, (__bridge CFStringRef)@"Product");

        if (name) {
            [array addObject:name];
        } else {
            [array addObject:@"noname"];
        }
    }

    return array;
}

#define IOHIDEventFieldBase(type) (type << 16)
#define kIOHIDEventTypeTemperature  15
#define kIOHIDEventTypePower        25

NSArray *getThermalValues(NSDictionary *sensors) {
    IOHIDEventSystemClientRef system = IOHIDEventSystemClientCreate(kCFAllocatorDefault);

    IOHIDEventSystemClientSetMatching(system, (__bridge CFDictionaryRef)sensors);
    NSArray *matchingsrvs = (__bridge NSArray *)IOHIDEventSystemClientCopyServices(system);

    long            count = [matchingsrvs count];
    NSMutableArray  *array = [[NSMutableArray alloc] init];

    for (int i = 0; i < count; i++) {
        IOHIDServiceClientRef   sc = (IOHIDServiceClientRef)matchingsrvs[i];
        IOHIDEventRef           event = IOHIDServiceClientCopyEvent(sc, kIOHIDEventTypeTemperature, 0, 0);

        NSNumber    *value;
        double      temp = 0.0;

        if (event != 0) {
            temp = IOHIDEventGetFloatValue(event, IOHIDEventFieldBase(kIOHIDEventTypeTemperature));
        }

        value = [NSNumber numberWithDouble:temp];
        [array addObject:value];
    }

    return array;
}

NSString *dumpNamesValues(NSArray *kvsN, NSArray *kvsV) {
    NSMutableString *valueString = [[NSMutableString alloc] init];
    int             count = [kvsN count];

    for (int i = 0; i < count; i++) {
        NSString *output = [NSString stringWithFormat:@"%s:%lf\n", [kvsN[i] UTF8String], [kvsV[i] doubleValue]];
        [valueString appendString:output];
    }

    return valueString;
}

char *getThermals() {
    NSDictionary    *thermalSensors = matching(0xff00, 5);
    NSArray         *thermalNames = getProductNames(thermalSensors);
    NSArray         *thermalValues = getThermalValues(thermalSensors);
    NSString        *result = dumpNamesValues(thermalNames, thermalValues);
    char            *finalStr = strdup([result UTF8String]);

    CFRelease(thermalSensors);
    CFRelease(thermalNames);
    CFRelease(thermalValues);
    CFRelease(result);

    return finalStr;
}
