/* Build-time replacements for the pinned OpenCV.js generated bindings.
 * These functions are injected into the upstream binding scope, where its type
 * converters and destructor helpers are available. No dynamic JS compilation is
 * needed; argument validation, conversions and cleanup match the generated code.
 */
function createNamedFunction(name, body) {
  return function () { 'use strict'; return body.apply(this, arguments); };
}

function makeDynCaller(dynCall) {
  return function () {
    var args = [rawFunction];
    for (var i = 1; i < signature.length; i++) args.push(arguments[i - 1]);
    return dynCall.apply(null, args);
  };
}

function craftInvokerFunction(humanName, argTypes, classType, cppInvokerFunc, cppTargetFunc) {
  var count = argTypes.length;
  if (count < 2) throwBindingError("argTypes array size mismatch! Must at least get return value and 'this' types!");
  var isMethod = argTypes[1] !== null && classType !== null;
  var needsStack = false;
  for (var i = 1; i < count; i++) {
    if (argTypes[i] !== null && argTypes[i].destructorFunction === undefined) needsStack = true;
  }
  return function () {
    if (arguments.length !== count - 2) {
      throwBindingError('function ' + humanName + ' called with ' + arguments.length + ' arguments, expected ' + (count - 2) + ' args!');
    }
    var destructors = needsStack ? [] : null;
    var wired = [cppTargetFunc];
    if (isMethod) wired.push(argTypes[1].toWireType(destructors, this));
    for (var i = 0; i < count - 2; i++) wired.push(argTypes[i + 2].toWireType(destructors, arguments[i]));
    var result = cppInvokerFunc.apply(null, wired);
    if (needsStack) {
      runDestructors(destructors);
    } else {
      for (var i = isMethod ? 1 : 2; i < count; i++) {
        if (argTypes[i].destructorFunction !== null) argTypes[i].destructorFunction(wired[isMethod ? i : i - 1]);
      }
    }
    if (argTypes[0].name !== 'void') return argTypes[0].fromWireType(result);
  };
}

function __emval_get_method_caller(argCount, argTypes) {
  var types = __emval_lookupTypes(argCount, argTypes);
  var returnType = types[0];
  return __emval_addMethodCaller(function (handle, name, destructors, pointer) {
    var args = [], offset = 0;
    for (var i = 1; i < argCount; i++) {
      args.push(types[i].readValueFromPointer(pointer + offset));
      offset += types[i].argPackAdvance;
    }
    var result = handle[name].apply(handle, args);
    for (var i = 1; i < argCount; i++) if (types[i].deleteObject) types[i].deleteObject(args[i - 1]);
    if (!returnType.isVoid) return returnType.toWireType(destructors, result);
  });
}
