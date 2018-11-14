package com.sap.piper

import java.lang.annotation.Retention
import java.lang.annotation.RetentionPolicy
import java.lang.annotation.Target
import java.lang.annotation.ElementType

@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.METHOD)
@interface Dependencies {
    String[] requiredPlugins() default []
    String[] optionalPlugin() default []
    String[] requiredShellCommands() default []
}
