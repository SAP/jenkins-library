package com.sap.piper

import java.lang.annotation.ElementType
import java.lang.annotation.Retention
import java.lang.annotation.RetentionPolicy
import java.lang.annotation.Target

/**
 * Methods or classes annotated with this annotation are used outside
 * this shared, e.g. in other shared libraries. In case there is the
 * need for changing this methods this should be clearly announced
 * in order to get a consensus about the change and in order to allow
 * users of the corresponding class/method to adapt to the change accordingly.
 */

@Retention(RetentionPolicy.RUNTIME)
@Target([ElementType.METHOD, ElementType.TYPE])
@interface API {
    /**
     * API marked as deprecated should not be used and moved to non-deprecated API.
     */
    boolean deprecated() default false
}
